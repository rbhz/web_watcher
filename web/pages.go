package web

const indexPageTemplate = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>HTTP checker</title>
  </head>
  <body>
      <div class="container">
          <div class="row">
                <table class="table">
                    <thead>
                        <tr>
                            <th scope="col">#</th>
                            <th scope="col">Url</th>
                            <th scope="col">Last change</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr class="d-none empty_row">
                            <th scope="row" class="num"></th>
                            <td class="url">
                                <a href=""></a>
                            </td>
                            <td class="change"></td>
                        </tr>
                    </tbody>
                </table>
          </div>
      </div>

    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.4.1/jquery.js" integrity="sha256-WpOohJOqMqqyKL9FccASB9O0KwACQJpFTUBLTYOVvVU=" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js" integrity="sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q" crossorigin="anonymous"></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" integrity="sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl" crossorigin="anonymous"></script>
    <script>
        $(document).ready(function() {
            $.get('/api/list', function(data) {
                let tbody = $('table tbody');
                data = JSON.parse(data);
                for (var idx = 0; idx < data.length; idx++) {
                    tbody.append($('.empty_row').clone());
                    row = $($('.empty_row')[1]);
                    row.attr('class', '');
                    row.find('.num').text(1 + idx);
                    row.find('.url a').text(data[idx].url).attr('href', data[idx].url);
                    let changed = new Date(data[idx].last_change);
                    row.find('.change').text(changed.toLocaleString());
                }

                ws = new WebSocket('ws://'+window.location.host+'/ws');
                ws.onopen = function(evt) {
                    console.log("ws OPEN");
                }
                ws.onclose = function(evt) {
                    console.log("ws CLOSE");
                }
                ws.onmessage = function(evt) {
                    data = JSON.parse(evt.data)
                    for (var idx = 0; idx < data.length; idx++) {
                        let item = data[idx];
                        let row = $('a[href="' +item.url+'"]').parents('tr');
                        changed = new Date(item.last_change);
                        row.find('.change').text(changed.toLocaleString());
                    }
                }
                ws.onerror = function(evt) {
                    console.log("ws ERROR: " + evt.data);
                }
            });
        });
    </script>
  </body>
</html>
`
