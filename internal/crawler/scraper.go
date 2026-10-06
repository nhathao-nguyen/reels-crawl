package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// ArticleContent holds the scraped article text and metadata
type ArticleContent struct {
	SourceURL   string    `json:"source_url"`
	ResolvedURL string    `json:"resolved_url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	MainText    string    `json:"main_text"`
	WordCount   int       `json:"word_count"`
	ScrapedAt   time.Time `json:"scraped_at"`
}

// ExtractLinksFromText finds all http/https links from text (captions, comments)
func ExtractLinksFromText(text string) []string {
	if text == "" {
		return nil
	}
	urlRegex := regexp.MustCompile(`https?://[^\s"'<>]+`)
	matches := urlRegex.FindAllString(text, -1)
	var valid []string
	seen := make(map[string]bool)
	for _, m := range matches {
		clean := strings.TrimRight(m, ".,);:!?")
		if !seen[clean] {
			seen[clean] = true
			valid = append(valid, clean)
		}
	}
	return valid
}

// CleanURL removes tracking parameters (fbclid, utm_*, etc.)
func CleanURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for key := range q {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "fbclid" || lower == "gclid" || lower == "ref" {
			q.Del(key)
		}
	}
	u.RawQuery = q.Encode()
	u.Fragment = "" // remove fragments like #google_vignette
	return u.String()
}

// ChapterContent represents a single scraped chapter
type ChapterContent struct {
	ChapterNum int    `json:"chapter_num"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	Text       string `json:"text"`
}

// ScrapeArticle resolves redirects and extracts the main readable text,
// optionally crawling sequential chapters if a chapter pattern is detected.
func ScrapeArticle(rawURL string) (*ArticleContent, error) {
	return ScrapeArticleWithChapters(rawURL, 20)
}

// ScrapeArticleWithChapters extracts article content, crawling up to maxChapters if chapter links exist.
func ScrapeArticleWithChapters(rawURL string, maxChapters int) (*ArticleContent, error) {
	if maxChapters <= 0 {
		maxChapters = 20
	}

	client := &http.Client{
		Timeout: 10 * time.Second, // 10s per request timeout as specified in BA
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("quá nhiều lần chuyển hướng (redirect loop)")
			}
			return nil
		},
	}

	cleanedURL := CleanURL(rawURL)
	firstArt, nextURL, err := fetchSinglePage(client, cleanedURL)
	if err != nil {
		return nil, err
	}

	firstArt.SourceURL = rawURL

	// Check if this article is a chapter
	chapterRegex := regexp.MustCompile(`(?i)(?:chapter|chap|part)[-_/]?(\d+)`)
	isChapter := chapterRegex.MatchString(cleanedURL) || chapterRegex.MatchString(firstArt.Title)

	if !isChapter || nextURL == "" {
		return firstArt, nil
	}

	// Sequential chapter crawling with circuit breaker
	var chapters []string
	chapters = append(chapters, firstArt.MainText)

	currURL := nextURL
	visited := map[string]bool{cleanedURL: true, currURL: true}

	for ch := 2; ch <= maxChapters && currURL != ""; ch++ {
		nextArt, nxt, err := fetchSinglePage(client, currURL)
		if err != nil || nextArt.MainText == "" {
			break
		}
		chapters = append(chapters, fmt.Sprintf("=== CHAPTER %02d: %s ===\n\n%s", ch, nextArt.Title, nextArt.MainText))

		if nxt == "" || visited[nxt] {
			break
		}
		visited[nxt] = true
		currURL = nxt
	}

	if len(chapters) > 1 {
		firstArt.MainText = strings.Join(chapters, "\n\n")
		words := strings.Fields(firstArt.MainText)
		firstArt.WordCount = len(words)
	}

	return firstArt, nil
}

func fetchSinglePage(client *http.Client, targetURL string) (*ArticleContent, string, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("không thể kết nối website: %w", err)
	}
	defer resp.Body.Close()

	resolvedURL := CleanURL(resp.Request.URL.String())

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("lỗi đọc cấu trúc HTML: %w", err)
	}

	var title, desc, nextChapterURL string
	var paragraphs []string

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)
			switch tag {
			case "script", "style", "nav", "header", "footer", "aside", "noscript", "svg", "button", "iframe":
				return
			case "title":
				if n.FirstChild != nil {
					title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "meta":
				var prop, name, content string
				for _, a := range n.Attr {
					if a.Key == "property" {
						prop = a.Val
					} else if a.Key == "name" {
						name = a.Val
					} else if a.Key == "content" {
						content = a.Val
					}
				}
				if prop == "og:title" && title == "" {
					title = content
				}
				if (prop == "og:description" || name == "description") && desc == "" {
					desc = content
				}
			case "a":
				linkText := strings.ToLower(extractNodeText(n))
				if strings.Contains(linkText, "next") || strings.Contains(linkText, "tiếp") || strings.Contains(linkText, "chương sau") {
					for _, a := range n.Attr {
						if a.Key == "href" && !strings.HasPrefix(a.Val, "#") && !strings.HasPrefix(a.Val, "javascript:") {
							nextChapterURL = resolveRelativeURL(resolvedURL, a.Val)
						}
					}
				}
			case "p", "article":
				text := extractNodeText(n)
				text = strings.TrimSpace(text)
				if len(text) > 30 {
					paragraphs = append(paragraphs, text)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	mainText := strings.Join(paragraphs, "\n\n")
	words := strings.Fields(mainText)

	return &ArticleContent{
		SourceURL:   targetURL,
		ResolvedURL: resolvedURL,
		Title:       title,
		Description: desc,
		MainText:    mainText,
		WordCount:   len(words),
		ScrapedAt:   time.Now(),
	}, nextChapterURL, nil
}

func resolveRelativeURL(baseRaw, relRaw string) string {
	baseURL, err := url.Parse(baseRaw)
	if err != nil {
		return relRaw
	}
	relURL, err := url.Parse(relRaw)
	if err != nil {
		return relRaw
	}
	return baseURL.ResolveReference(relURL).String()
}

// SaveArticleToDirectory saves article files to the target content directory.
// If chapters were scraped, saves both chapters/ and full_article.txt according to BA specs.
func SaveArticleToDirectory(article *ArticleContent, targetDir string) error {
	contentDir := filepath.Join(targetDir, "content")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return err
	}

	// Save summary / metadata JSON
	summaryPath := filepath.Join(contentDir, "summary.json")
	summaryData, _ := json.MarshalIndent(article, "", "  ")
	_ = os.WriteFile(summaryPath, summaryData, 0644)

	// Save full_article.txt
	txtPath := filepath.Join(contentDir, "full_article.txt")
	_ = os.WriteFile(txtPath, []byte(article.MainText), 0644)

	// If article contains chapters, save individual chapter files
	if strings.Contains(article.MainText, "=== CHAPTER ") {
		chaptersDir := filepath.Join(contentDir, "chapters")
		_ = os.MkdirAll(chaptersDir, 0755)

		sections := strings.Split(article.MainText, "=== CHAPTER ")
		for i, sec := range sections {
			trimmed := strings.TrimSpace(sec)
			if trimmed == "" {
				continue
			}
			chFile := filepath.Join(chaptersDir, fmt.Sprintf("chapter_%02d.txt", i))
			_ = os.WriteFile(chFile, []byte("=== CHAPTER "+trimmed), 0644)
		}
		// Save combined chapter file
		chFullFile := filepath.Join(chaptersDir, "full_article.txt")
		_ = os.WriteFile(chFullFile, []byte(article.MainText), 0644)
	}

	return nil
}

func extractNodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(extractNodeText(c))
	}
	return sb.String()
}
