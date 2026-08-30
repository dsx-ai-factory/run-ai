const _escape = document.createElement("textarea");

function escapeHTML(html: string) {
  _escape.textContent = html;
  return _escape.innerHTML;
}

export function formatMarkdown(text: string) {
  return escapeHTML(text)
    .replace(/ /g, "&nbsp;")
    .replace(/\*(.+?)\*/g, "<b>$1</b>")
    .replace(/\[(.+?)\]\((http.+?)\)/g, "<a target='_blank' href='$2'>$1</a>")
    .replace(/\[(.+?)\]\((.+?)\)/g, "<a href='$2'>$1</a>");
}
