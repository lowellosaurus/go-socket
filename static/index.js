// TODO: Actually figure out whey we wrap this in an anonymous function.
// Using an example pulled from MDN: https://developer.mozilla.org/en-US/docs/Web/Guide/AJAX/Getting_Started
(function() {
  var httpRequest;
  var id = "";

  // On load:
  addParticipant();

  // Add participant
  function addParticipant() {
    httpRequest = new XMLHttpRequest();
    httpRequest.onreadystatechange = handleAddParticipantResponse;
    httpRequest.open('GET', 'add_participant');
    httpRequest.send();
  }

  function handleAddParticipantResponse() {
    if (httpRequest.readyState === XMLHttpRequest.DONE) {
      if (httpRequest.status === 200) {
        var resp = JSON.parse(httpRequest.responseText);
        
        document.getElementById("my-id").innerHTML = resp.Id;
        updatePointsList(resp.Points);

        var ws = startWebsocket(resp.Id);
        document.getElementById("increase-score").addEventListener('click', increaseScore(resp.Id));
        window.addEventListener('beforeunload', closeWebsocket(ws, resp.Id));
      } else {
        alert("There was a problem with the request.");
      }
    }
  }

  // Increase score
  function increaseScore(req_id) {
    return function() {
      httpRequest = new XMLHttpRequest();
      httpRequest.onreadystatechange = handleIncreaseScoreResponse;
      httpRequest.open('GET', 'increment_score?id='+req_id);
      httpRequest.send();
    };
  }

  function handleIncreaseScoreResponse() {
    if (httpRequest.readyState === XMLHttpRequest.DONE) {
      if (httpRequest.status === 200) {
        // Rely on the websocket to update the DOM.
      } else {
        console.log("There was a problem with the request.");
      }
    }
  }

  function startWebsocket(ws_id) {
    ws = new WebSocket("ws://localhost:8888/websocket?id="+ws_id);
    
    ws.onmessage = function(evt) {
      var resp = JSON.parse(evt.data);
      document.getElementById("my-id").innerHTML = resp.Id;
      updatePointsList(resp.Points);
    };

    return ws;
  }

  function closeWebsocket(ws, req_id) {
    return function(evt) {
      evt.preventDefault();

      httpRequest = new XMLHttpRequest();
      httpRequest.onreadystatechange = handleCloseWebsocketResponse;
      httpRequest.open('GET', 'remove_participant?id='+req_id);
      httpRequest.send();

      evt.returnValue = '';
    };
  }

  function handleCloseWebsocketResponse() {
    if (httpRequest.readyState === XMLHttpRequest.DONE) {
      if (httpRequest.status === 200) {
        window.close();
      } else {
        console.log("There was a problem with the request.");
      }
    }
  }

  function updatePointsList(pts) {
    var listElements = "";
    for (let p of pts) {
      listElements += "<li>" + p + "</li>";
    }

    document.getElementById("points").innerHTML = listElements;
  }
})();
