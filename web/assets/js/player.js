$(document).ready(function () {
  var id = 1
  var total = 0
  const playerEl = $('#player')
  const prevEl = $('#prev')
  const nextEl = $('#next')
  /*
  const logoutEl = $('#logout')
  */

  prevEl.click(function () {
    if (id > 1) {
      id--
      playerEl.attr('src', `/api/mp4/${id}`)
    }
  })

  nextEl.click(function () {
    if (id < total) {
      id++
      playerEl.attr('src', `/api/mp4/${id}`)
    }
  })
  /*
  logoutEl.click(function () {
    if (confirm('确定要退出吗？')) {
      $.ajax({
        url: '/api/logout',
        method: 'POST',
        success: function () {
          window.location.href = '/login.html'
        },
        error: function (xhr, status, error) {
          alert('logout failed: ' + error)
        },
      })
    }
  })
  */
  $.ajax({
    url: '/api/mp4/total',
    method: 'GET',
    success: function (data) {
      total = data.data
      const params = new URLSearchParams(window.location.search)
      if (params.has('id')) {
        var param_id = parseInt(params.get('id'))
        if (param_id >= 1 && param_id <= total) {
          id = param_id
        }
      }
      playerEl.attr('src', `/api/mp4/${id}`)
    },
    error: function (xhr, status, error) {
      window.location.href = '/login.html'
    },
  })
})
