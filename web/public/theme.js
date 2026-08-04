// Applies the persisted theme before first paint. External (not inline) so the
// app's Content-Security-Policy can use script-src 'self' with no 'unsafe-inline'.
if (localStorage.getItem('theme') !== 'light') {
  document.documentElement.classList.add('dark');
}
