// Add a "Copy" button to each Go starter / code block.
document.querySelectorAll('pre.code').forEach((pre) => {
  const btn = document.createElement('button');
  btn.textContent = 'Copy';
  btn.className = 'copy-btn';
  btn.addEventListener('click', () => {
    navigator.clipboard.writeText(pre.innerText).then(() => {
      btn.textContent = 'Copied!';
      setTimeout(() => (btn.textContent = 'Copy'), 1200);
    });
  });
  pre.style.position = 'relative';
  btn.style.cssText =
    'position:absolute;top:8px;right:8px;font-size:12px;padding:2px 8px;' +
    'background:#2c3645;color:#e6e6e6;border:1px solid #3a4655;border-radius:6px;cursor:pointer;';
  pre.appendChild(btn);
});
