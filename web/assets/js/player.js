$(document).ready(function () {
  var id = 1
  var total = 0
  const playerEl = $('#player')
  const prevEl = $('#prev')
  const nextEl = $('#next')
  const currEl = $('#curr')
  const totalEl = $('#total')
  const logoutEl = $('#logout')

  prevEl.click(function () {
    if (id > 1) {
      id--
      playerEl.attr('src', `/api/mp4/${id}`)
      currEl.val(id)
    }
  })

  nextEl.click(function () {
    if (id < total) {
      id++
      playerEl.attr('src', `/api/mp4/${id}`)
      currEl.val(id)
    }
  })

  currEl.change(function () {
    const value = parseInt(currEl.val())
    if (value >= 1 && value <= total && value !== id) {
      id = value
      playerEl.attr('src', `/api/mp4/${id}`)
    } else {
      currEl.val(id)
    }
  })

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

  $.ajax({
    url: '/api/mp4/total',
    method: 'GET',
    success: function (data) {
      total = data.data
      totalEl.text(total)
      const params = new URLSearchParams(window.location.search)
      if (params.has('id')) {
        var param_id = parseInt(params.get('id'))
        if (param_id >= 1 && param_id <= total) {
          id = param_id
        }
      }
      currEl.val(id)
      playerEl.attr('src', `/api/mp4/${id}`)
    },
    error: function (xhr, status, error) {
      window.location.href = '/login.html'
    },
  })
})
