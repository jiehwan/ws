// server.js

"use strict";

const WebSocket = require("ws");

const wss = new WebSocket.Server({ 
	port: 4000 
});

console.log('port = ', wss.options.port);

var count_tx = 0, count_rx = 0;

function timeout(ws) {
        setTimeout(function setTimeout() {
                ws.send(JSON.stringify({ msg2: 'message', 
					count_tx: count_tx, 
					count_rx: count_rx}),
			function (err) {
				if (err) {
					console.log(err);
				}
				else {
					count_tx = count_tx + 1;
					console.info('tx[%d] rx[%d]', count_tx, count_rx);
					timeout(ws);
				}
			});
                }, 1000)
        }

wss.on("connection", function connection(ws) {
        console.info("websocket connection open");

	ws.on('message', function incoming(message) {
		count_rx = count_rx + 1;
		console.log('received: %s', message);
		});

        if (ws.readyState === ws.OPEN) {
                timeout(ws);
                }
        });
