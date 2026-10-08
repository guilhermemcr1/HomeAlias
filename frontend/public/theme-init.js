try {
  var t = localStorage.getItem('homealias-theme')
  if (t !== 'homealias-dark' && t !== 'homealias-light') {
    t = matchMedia('(prefers-color-scheme: light)').matches ? 'homealias-light' : 'homealias-dark'
  }
  document.documentElement.setAttribute('data-theme', t)
} catch (e) {}
