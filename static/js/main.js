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

  // Confirm destructive actions: data-confirm on a button (including submit buttons).
  document.querySelectorAll('[data-confirm]').forEach(el => {
    el.addEventListener('click', e => {
      if (!confirm(el.dataset.confirm)) e.preventDefault();
    });
  });

  // ── Per-table text filter ───────────────────────────────────────────────
  // Wire up any input with data-table-filter="<table-id>" to filter that table.
  document.querySelectorAll('[data-table-filter]').forEach(input => {
    const table = document.getElementById(input.dataset.tableFilter);
    if (!table) return;
    input.addEventListener('input', () => {
      const q = input.value.toLowerCase();
      table.querySelectorAll('tbody tr').forEach(row => {
        row.style.display = row.textContent.toLowerCase().includes(q) ? '' : 'none';
      });
    });
  });

  // ── Sortable table columns ──────────────────────────────────────────────
  // Tables with class table-sortable get click-to-sort on header cells that
  // have text content (the empty actions column is skipped automatically).
  document.querySelectorAll('table.table-sortable').forEach(table => {
    const ths = Array.from(table.querySelectorAll('thead th'));
    ths.forEach((th, colIdx) => {
      if (!th.textContent.trim()) return; // skip the actions column

      th.style.cursor = 'pointer';
      th.setAttribute('title', 'Click to sort');

      th.addEventListener('click', () => {
        const tbody = table.querySelector('tbody');
        // Skip the "No records" colspan row.
        const rows = Array.from(tbody.querySelectorAll('tr'))
          .filter(r => r.querySelectorAll('td').length > 1);

        const asc = th.dataset.sortDir !== 'asc';

        // Reset indicators on all headers.
        ths.forEach(t => {
          t.dataset.sortDir = '';
          const icon = t.querySelector('.sort-icon');
          if (icon) icon.remove();
        });

        th.dataset.sortDir = asc ? 'asc' : 'desc';
        const icon = document.createElement('span');
        icon.className = 'sort-icon ms-1 small opacity-75';
        icon.textContent = asc ? '▲' : '▼';
        th.appendChild(icon);

        rows.sort((a, b) => {
          const ca = a.querySelectorAll('td')[colIdx]?.textContent.trim() ?? '';
          const cb = b.querySelectorAll('td')[colIdx]?.textContent.trim() ?? '';
          // Try numeric comparison first.
          const na = parseFloat(ca.replace(/[^0-9.-]/g, ''));
          const nb = parseFloat(cb.replace(/[^0-9.-]/g, ''));
          const cmp = !isNaN(na) && !isNaN(nb) ? na - nb : ca.localeCompare(cb);
          return asc ? cmp : -cmp;
        });

        rows.forEach(r => tbody.appendChild(r));
      });
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
