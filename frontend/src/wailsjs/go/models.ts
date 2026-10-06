export namespace config {
	
	export class Config {
	    app_mode: string;
	    capcut_draft_path: string;
	    template_name: string;
	    whisper_provider: string;
	    groq_api_key: string;
	    local_whisper_model: string;
	    export_resolution: string;
	    download_dir: string;
	    max_chapters_per_story: number;
	    request_delay_seconds: number;
	    auto_delete_raw: boolean;
	    skip_duplicates: boolean;
	    extract_comment_links: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_mode = source["app_mode"];
	        this.capcut_draft_path = source["capcut_draft_path"];
	        this.template_name = source["template_name"];
	        this.whisper_provider = source["whisper_provider"];
	        this.groq_api_key = source["groq_api_key"];
	        this.local_whisper_model = source["local_whisper_model"];
	        this.export_resolution = source["export_resolution"];
	        this.download_dir = source["download_dir"];
	        this.max_chapters_per_story = source["max_chapters_per_story"];
	        this.request_delay_seconds = source["request_delay_seconds"];
	        this.auto_delete_raw = source["auto_delete_raw"];
	        this.skip_duplicates = source["skip_duplicates"];
	        this.extract_comment_links = source["extract_comment_links"];
	    }
	}

}

export namespace database {
	
	export class PipelineStats {
	    total: number;
	    pending: number;
	    downloaded: number;
	    draft_created: number;
	    exported: number;
	    failed: number;
	
	    static createFrom(source: any = {}) {
	        return new PipelineStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.pending = source["pending"];
	        this.downloaded = source["downloaded"];
	        this.draft_created = source["draft_created"];
	        this.exported = source["exported"];
	        this.failed = source["failed"];
	    }
	}
	export class TrackedSource {
	    id: number;
	    platform: string;
	    source_url: string;
	    author_name: string;
	    // Go type: time
	    last_checked_at?: any;
	    video_count?: number;
	
	    static createFrom(source: any = {}) {
	        return new TrackedSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.platform = source["platform"];
	        this.source_url = source["source_url"];
	        this.author_name = source["author_name"];
	        this.last_checked_at = this.convertValues(source["last_checked_at"], null);
	        this.video_count = source["video_count"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Video {
	    video_id: string;
	    source_id?: number;
	    platform: string;
	    title: string;
	    template_name: string;
	    status: string;
	    export_path: string;
	    error_message: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Video(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.video_id = source["video_id"];
	        this.source_id = source["source_id"];
	        this.platform = source["platform"];
	        this.title = source["title"];
	        this.template_name = source["template_name"];
	        this.status = source["status"];
	        this.export_path = source["export_path"];
	        this.error_message = source["error_message"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

