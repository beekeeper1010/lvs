$(document).ready(function () {
  $.ajax({
    url: '/api/mp4/list',
    method: 'GET',
    success: function (data) {
      var container = $('#mp4-list')
      data.data.forEach(function (item) {
        var card = $('<div class="card"></div>')
        var thumbnail = $(
          '<img class="thumbnail" src="' + item.thumbnail + '" alt="缩略图">'
        )
        var info = $('<div class="info"></div>')
        var name = item.name
        if (name.length > 24) {
          name = name.substring(0, 24) + '...'
        }
        var minutes = Math.floor(item.duration / 60)
        var seconds = item.duration % 60
        var desc = $(`<p>${name} (${minutes}分${seconds}秒)</p>`)
        info.append(desc)
        card.append(thumbnail, info)
        container.append(card)

        card.click(function () {
          window.open(`player.html?id=${item.ID}`, '_blank')
        })
      })
    },
    error: function () {
      window.location.href = '/login.html'
    },
  })
})
