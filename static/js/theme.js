function initTheme() {
  const savedTheme = localStorage.getItem('theme') || 'light';
  document.body.setAttribute('data-theme', savedTheme);
  updateThemeToggle(savedTheme === 'dark');
}

function toggleTheme() {
  const currentTheme = document.body.getAttribute('data-theme');
  const newTheme = currentTheme === 'dark' ? 'light' : 'dark';

  document.body.setAttribute('data-theme', newTheme);
  localStorage.setItem('theme', newTheme);
  updateThemeToggle(newTheme === 'dark');
}

function updateThemeToggle(isDark) {
  const toggleButton = document.getElementById('theme-toggle');
  if (toggleButton) {
      toggleButton.textContent = isDark ? '☀️ Light Mode' : '🌙 Dark Mode';
  }
}

document.addEventListener('DOMContentLoaded', initTheme);
