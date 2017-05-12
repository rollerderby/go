function websocket(service, options) {
	var ws = {
		options: {
			hostname: location.hostname,
			port: location.port,
			debug: false,
			autoReconnect: true,
			pingInterval: 8 * 60 * 1000,
			reconnectInterval: 500,
		},
		socket: null,
		callbacks: new Array(),
		reconnecting: false,
		url: null,

		Init: function(url, opt) {
			for (var name in opt) {
				ws.options[name] = opt[name];
			}
			ws.url = url;

			setInterval(ws._ping, ws.options.pingInterval);
			setInterval(ws._reconnect, ws.options.reconnectInterval);
			ws._reset();
		},

		_connect: function() {
			if (ws.socket != null)
				return;
			var proto = "ws://";
			if (ws.options.secure) {
				proto = "wss://";
			}
			var base_url = proto + ws.options.hostname + (ws.options.port ? ':' + ws.options.port : '');
			ws.socket = ws._makeSocket(base_url + ws.url);
		},

		_socketIdx: 0,
		_makeSocket: function(url) {
			var sock = {
				socket: null,
				socketIdx: ++ws._socketIdx,

				connect: function(url) {
					if (ws.options.debug) {
						console.log("ws._makeSocket.connect", url);
					}
					sock.socket = new WebSocket(url);
					sock.socket.onopen = sock.onopen;
					sock.socket.onclose = sock.onclose;
					sock.socket.onmessage = sock.onmessage;
					sock.socket.onerror = sock.onerror;
				},

				onopen: function(e) {
					if (ws.socket != null && ws.socket.socketIdx == sock.socketIdx) {
						return ws._onOpen(e);
					} else {
						// Oops!  Not the active socket, close it and continue on
						sock.socket.onopen = function() {};
						sock.socket.onclose = function() {};
						sock.socket.onerror = function() {};
						sock.socket.onmessage = function() {};
						sock.socket.close();
					}
				},

				onclose: function(e) {
					if (ws.socket != null && ws.socket.socketIdx == sock.socketIdx) {
						return ws._onClose(e);
					} else {
						// Oops!  Not the active socket
						sock.socket.onopen = function() {};
						sock.socket.onclose = function() {};
						sock.socket.onerror = function() {};
						sock.socket.onmessage = function() {};
					}
				},

				onmessage: function(e) {
					if (ws.socket != null && ws.socket.socketIdx == sock.socketIdx) {
						return ws._onMessage(e);
					} else {
						// Oops!  Not the active socket, close it and continue on
						sock.socket.onopen = function() {};
						sock.socket.onclose = function() {};
						sock.socket.onerror = function() {};
						sock.socket.onmessage = function() {};
						sock.socket.close();
					}
				},

				onerror: function(e) {
					if (ws.socket != null && ws.socket.socketIdx == sock.socketIdx) {
						return ws._onError(e);
					} else {
						// Oops!  Not the active socket, close it and continue on
						sock.socket.onopen = function() {};
						sock.socket.onclose = function() {};
						sock.socket.onerror = function() {};
						sock.socket.onmessage = function() {};
						sock.socket.close();
					}
				},
			}

			sock.connect(url);
			return sock;
		},

		_onOpen: function(e) {
			if (ws.options.debug) {
				console.log("ws._onOpen", e);
			}
			ws.reconnecting = false;
			if (ws.options.onOpen != null) {
				ws.options.onOpen(e);
			}
		},

		_onClose: function(e) {
			if (ws.options.debug) {
				console.log("ws._onClose", e);
			}
			ws.reconnecting = false;
			ws._reset();
			if (ws.options.onClose != null) {
				ws.options.onClose(e);
			}
		},

		_onMessage: function(e) {
			if (ws.options.debug) {
				console.log("ws._onMessage", e);
			}
			ws.reconnecting = false;
			ws.processCallback(e);
			if (ws.options.onMessage != null) {
				ws.options.onMessage(e);
			}
		},

		_onError: function(e) {
			if (ws.options.debug) {
				console.log("ws._onError", e);
			}
			ws.reconnecting = false;
			ws._reset();
			if (ws.options.onError != null) {
				ws.options.onError(e);
			}
		},

		Send: function(type, data) {
			try {
				var jObj = {
					type: type,
					data: data
				}
				var json = JSON.stringify(jObj);
				if (ws.options.debug) {
					console.log("ws.Send", type, json);
				}
				ws.socket.socket.send(json);
			} catch (e) {
				if (ws.options.onError != null) {
					ws.options.onError(e);
				}
				ws._reset();
			}
		},

		processCallback: function(e) {
			var obj;
			try {
				obj = JSON.parse(e.data);
			} catch (err) {
				console.log("cannot parse JSON: ", err, e);
				return;
			}
			if (ws.options.debug)
				console.log("ws.processCallback", obj);
			if (obj.type != null) {
				var t = obj.type.toLowerCase();
				var handled = false;

				if (t == "pong") {
					handled = true;
				} else {
					var arrayLength = ws.callbacks.length;
					for (var i = 0; i < arrayLength; i++) {
						var callback = ws.callbacks[i];
						if (callback.type == t) {
							handled = true
							callback.func(obj);
						}
					}
				}
				if (!handled) {
					console.log("Cannot find a callback for " + obj.type);
				}
			}
		},

		_reset: function() {
			if (ws.options.debug) {
				console.log("ws._reset");
			}
			if (ws.socket == null) {
				ws._connect();
				return;
			}

			if (ws.options.onClose != null) {
				ws.options.onClose();
			}
			ws.socket = null;
		},

		_reconnect: function() {
			if (ws.socket != null) {
				return
			}

			if (ws.options.autoReconnect && !ws.reconnecting) {
				ws.reconnecting = true;
				if (ws.options.debug) {
					console.log("Reconnecting", ws.socket);
				}
				ws._connect();
			}
		},

		_ping: function() {
			if (ws.socket == null) {
				return;
			}

			ws.Send("ping");
		},

		Register: function(type, func) {
			ws.callbacks.push({
				type: type.toLowerCase(),
				func: func
			});
		},
	}

	ws.Init(service, options);
	return ws
}

var modalDiv = null;
var control = null;

function modalClose() {
	if (modalDiv != null) {
		modalDiv.remove();
		modalDiv = null;
	}
}

function modal(headerText, div, allowClose) {
	modalClose();

	modalDiv = $("<div>").addClass("modal");
	var content = $("<div>").addClass("modal-content").appendTo(modalDiv);
	var header = $("<div>").addClass("modal-header").appendTo(content);
	if (allowClose) {
		var span = $("<span>").addClass("close").html("&times;").appendTo(header);
		span.on('click', modalClose);
	}
	$("<h2>").text(headerText).appendTo(header);
	div.addClass("modal-body").appendTo(content);

	console.log(modalDiv);
	modalDiv.appendTo(document.body);
}

function newControlChannel() {
	var cc = {
		ws: null,
		clear: function() {},

		menuItems: function(msg) {
			console.log("menuItems", msg);
			if (msg.data != null) {
				var nav = $("div#topMenu ul.navigation");
				var index = $("ul.menu.index");
				nav.find("li:not(.sub-heading)").detach();
				index.empty();
				for (var i = 0; i < msg.data.length; i++) {
					var item = msg.data[i];
					if (item.Priority == 1) {
						$("div#topMenu li.auth.link").removeClass("Show");
						$("div#topMenu li.auth.link." + item.Display).addClass("Show");
					} else if (item.Priority > 1) {
						nav.append($("<li>").append($("<a>").addClass("menuEntry").prop("href", item.Path).text(item.Display)));
					}
					index.append($("<li>").append($("<a>").prop("href", item.Path).text(item.Display)));
				}
			}
		},

		user: function(msg) {
			console.log("user", msg);
		},

		reauth: function() {
			document.location = "/auth/?reauth";
		},

		Init: function() {
			cc.ws = websocket("/ws/control", {
				onError: cc.clear,
				onClose: cc.clear
			});
			cc.ws.Register("User", cc.user);
			cc.ws.Register("MenuItems", cc.menuItems);
			cc.ws.Register("Reauth", cc.reauth);
		},
	};

	return cc;
}

function setupControlChannel() {
	control = newControlChannel();
	control.Init();
}

$().ready(setupControlChannel);
