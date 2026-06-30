/* webui — application JS (Bootstrap handles dropdowns and navbar) */

document.addEventListener('DOMContentLoaded', () => {
  // Mark the active nav link by matching the current path prefix.
  const path = window.location.pathname;
  document.querySelectorAll('.navbar .nav-link[href]').forEach(a => {
    const href = a.getAttribute('href');
    if (href && href !== '/' && path.startsWith(href)) {
      a.classList.add('active');
    } else if (href === path) {
      a.classList.add('active');
    }
  });

  // Auto-dismiss alerts after 4 s.
  document.querySelectorAll('.alert[data-auto-dismiss]').forEach(el => {
    setTimeout(() => el.remove(), 4000);
  });

  // Confirm destructive actions on elements with data-confirm.
  document.querySelectorAll('[data-confirm]').forEach(el => {
    el.addEventListener('click', e => {
      if (!confirm(el.dataset.confirm)) e.preventDefault();
    });
  });
});

/**
 * stopServer — called by the Stop button in the top bar.
 * POSTs to /api/shutdown, then tries to close the tab.
 * Falls back to a full-page overlay if window.close() is blocked.
 */
async function stopServer() {
  const btn = document.querySelector('.btn-stop, [onclick="stopServer()"]');
  if (btn) {
    btn.disabled = true;
    btn.textContent = 'Stopping…';
  }

  try {
    await fetch('/api/shutdown', { method: 'POST' });
  } catch (_) {
    // Server closed the connection before responding — expected.
  }

  window.close();

  // Fallback overlay if the browser refused to close the tab.
  const overlay = document.createElement('div');
  overlay.className = 'stopped-overlay';
  overlay.innerHTML = `
    <div class="stopped-box">
      <h2>Server stopped</h2>
      <p>You may close this tab.</p>
    </div>`;
  document.body.appendChild(overlay);
}
