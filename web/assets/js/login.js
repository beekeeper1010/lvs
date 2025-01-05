$(document).ready(function () {
  $(document).on('mousemove', function (event) {
    var dw = $(document).width() / 15
    var dh = $(document).height() / 15
    var x = event.pageX / dw
    var y = event.pageY / dh
    $('.eye-ball').css({
      width: x,
      height: y,
    })
  })

  $('.btn').click(function (event) {
    event.preventDefault()
    const username = $('.form-control').eq(0).val()
    const password = $('#password').val()
    $.ajax({
      url: '/api/login',
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({ username, password }),
      success: function (response) {
        window.location.href = '/plaza.html'
      },
      error: function (xhr, status, error) {
        $('form').addClass('wrong-entry')
        setTimeout(function () {
          $('form').removeClass('wrong-entry')
        }, 2000)
      },
    })
  })
})
