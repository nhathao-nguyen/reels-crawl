// Lightweight declarative DOM builder helper (Hyperscript style)
// Allows creating 100% pure DOM elements in JavaScript without writing raw HTML.

export function el(tag, props = {}, ...children) {
  const element = document.createElement(tag);

  for (const [key, val] of Object.entries(props)) {
    if (val === undefined || val === null) continue;

    if (key === 'className') {
      element.className = val;
    } else if (key === 'style' && typeof val === 'object') {
      Object.assign(element.style, val);
    } else if (key.startsWith('on') && typeof val === 'function') {
      element.addEventListener(key.slice(2).toLowerCase(), val);
    } else if (key === 'dataset' && typeof val === 'object') {
      Object.assign(element.dataset, val);
    } else if (key === 'checked' || key === 'value' || key === 'disabled') {
      element[key] = val;
    } else {
      element.setAttribute(key, val);
    }
  }

  for (const child of children.flat(Infinity)) {
    if (child === null || child === undefined || child === false) continue;
    if (typeof child === 'string' || typeof child === 'number') {
      element.appendChild(document.createTextNode(String(child)));
    } else if (child instanceof Node) {
      element.appendChild(child);
    }
  }

  return element;
}

export function clear(element) {
  while (element.firstChild) {
    element.removeChild(element.firstChild);
  }
}
