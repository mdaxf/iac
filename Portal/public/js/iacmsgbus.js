var IACMessageClient = function (server){
	const HubPath = "/iacmessagebus";
	const HubName = "IACMessageBusHub";
	if (IACMessageBus && IACMessageBus.connection && IACMessageBus.connection._connectionState === "Connected")
		return IACMessageBus;
	
	var _client = this;
	IACMessageBus = _client
	
	this.subscribers ={};
	this.CallbackMap = {};
	this.initialized = false;
	this.autoConnect = true;
	this.connection = null;
	this.Queue =[];
	this.disconnectHandlerMap = [];
	this.connectionID = "";
	
	if (!server || server =="")
		server = window.location
	
	var serverUrl = server + HubPath

	this.Connect = function(){
		_client.connection = new signalR.HubConnectionBuilder()
			.withUrl(serverUrl, {
				withCredentials: false
			})
			.build();
				
		_client.connection.start().then(function () {
			_client.initialized = true;	
			
			_client.connectionID = 	_client.connection.connection.connectionId
	
			for (var idx = 0; idx < _client.Queue.length; idx++)
			{
				var call = _client.Queue[idx];
				call[0].apply(null, call[1]);
			}
			
			var topics = Object.keys(_client.CallbackMap);
			
			console.log(topics, _client)
			
			for (var idx = 0; idx < topics.length; idx++)
			{				
				if(_client.CallbackMap.hasOwnProperty(topics[idx]))
					if (_client.CallbackMap[topics[idx]].length > 0)
						_client.connection.on(topics[idx], message => {
							console.log("receive message for topic:", topic, message)
							_client.executesubcallback(topics[idx], message);
						});
			}
		});	

		console.log(_client, _client.connection)
		
	}
	
	this.Subscribe = function  (topic, callback) {
		if (!_client.initialized)
		{
			_client.Queue.push([_client.Subscribe, [topic, callback]]);
			return;
		}
		
		
		if (!_client.CallbackMap.hasOwnProperty(topic))
		{
			_client.CallbackMap[topic] = [];
			_client.CallbackMap[topic].push(callback);
			_client.connection.on(topic, message => {
				console.log("receive message for topic:", topic, message)
				_client.executesubcallback(topic,message);
			});
			
		}else{			
			_client.CallbackMap[topic].push(callback);
		}		
	}
	
	this.executesubcallback = function(topic,message){
		console.log("execute the sub callback: ", topic, message)
		if(!_client.CallbackMap.hasOwnProperty(topic))
			return;
		
		var callbacks = _client.CallbackMap[topic];
		for(var idx =0; idx<callbacks.length ; idx++ ){
			console.log("execute: ")
			console.log(callbacks[idx], message)			
			callbacks[idx](message)
		}
	}
	
	this.Unsubscribe = function  (topic, callback) {
		if (!_client.initialized)
		{
			_client.Queue.push([_client.Unsubscribe, [topic]]);
			return;
		}		
		
		if (_client.CallbackMap.hasOwnProperty(topic))
		{
			_client.connection.off(topic, callback);
			
			var callbacksString = _client.CallbackMap[topic].map(function (val, idx) {return '' + val;})
			var idx = $.inArray(''+callback, callbacksString);
			if (idx >= 0)
				_client.CallbackMap[topic].splice(idx, 1);
			
			if (_client.CallbackMap[topic].length < 1)
			{
				_client.CallbackMap[topic] = null;
				delete _client.CallbackMap[topic];
				
			}
		}		
	}

	this.Publish = function (topic,message) {
		_client.connection.invoke("send", topic,message,_client.connectionID)
		
	}
	
	this.Broadcast = function(message) {
		_client.connection.invoke("broadcast", message,_client.connectionID);
	}
	
	this.Echo = function(message) {
		_client.connection.invoke("echo", message,_client.connectionID);
	}
	
		
	this.AddDisconnectHandler = function (handler){
		_client.disconnectHandlerMap.push(handler); 
	}
	
	this.RemoveDisconnectHandler = function (handler){
		var callbacksString = _client.disconnectHandlerMap.map(function (val, idx) {return '' + val;})
		var idx = $.inArray(''+handler, callbacksString);
		if (idx >= 0)
			_client.disconnectHandlerMap.splice(idx, 1);
	}
	
	this.SetAutoConnect = function (autoConnect) {
		_client.autoConnect = autoConnect;
		if (_client.autoConnect)
			_client.connection.start()
	}
		
	this.Disconnect = function () {
		_client.autoConnect = false;
		_client.connection.stop()
	}
	
	window.addEventListener('beforeunload', _client.Disconnect);
	this.Connect()
}
var IACMessageBus;

/*

var IACMessageClient1 = function ($Context) {
	const HubPath = "/iacmessagebus";
	const HubName = "IACMessageBusHub";
	
	
	if (IACMessageBus && IACMessageBus.hubConnection && IACMessageBus.hubConnection.state === 1)
		return IACMessageBus;
	
	var _client = this;
	IACMessageBus = _client;

	this.$Context = $Context;
	this.hubConnection = $.hubConnection("", { useDefaultPath: false, logging: true });
	var getUrl = window.location;
	var hubUrl = "http://127.0.0.1:8222";
	_client.hubConnection.url = hubUrl + HubPath;
	this.proxy = _client.hubConnection.createHubProxy(HubName);
	this.CallbackMap = {};
	this.initialized = false;
	this.autoConnect = true;
	this.Queue =[];
	this.disconnectHandlerMap = [];
	
	this.S4 = function () {
		return (((1+Math.random())*0x10000)|0).toString(16).substring(1); 
	}
	
	this.ClientID = (_client.S4() + _client.S4() + "-" + _client.S4() + "-4" + _client.S4().substr(0,3) + "-" + _client.S4() + "-" + _client.S4() + _client.S4() + _client.S4()).toLowerCase();
	
	this.Publish = function (topic, message) {
		if (!_client.initialized)
		{
			_client.Queue.push([_client.Publish, [topic, message]]);
			return;
		}
		
		if (typeof message != 'string')
			message = JSON.stringify(message);
		_client.proxy.invoke('send', topic, message, _client.ClientID);
	};
	
	this.Subscribe = function  (topic, callback) {
		if (!_client.initialized)
		{
			_client.Queue.push([_client.Subscribe, [topic, callback]]);
			return;
		}
				
		if (_client.CallbackMap[topic] == null)
		{
			_client.proxy.invoke('subscribe', topic);
			_client.CallbackMap[topic] = [];
		}			
		_client.CallbackMap[topic].push(callback);
	}
		
	this.Unsubscribe = function  (topic, callback) {
		if (!_client.initialized)
		{
			_client.Queue.push([_client.Unsubscribe, [topic]]);
			return;
		}		
		
		if (_client.CallbackMap[topic] != null)
		{
			var callbacksString = _client.CallbackMap[topic].map(function (val, idx) {return '' + val;})
			var idx = $.inArray(''+callback, callbacksString);
			if (idx >= 0)
				_client.CallbackMap[topic].splice(idx, 1);
			
			if (_client.CallbackMap[topic].length < 1)
			{
				_client.proxy.invoke('unsubscribe', topic);
				_client.CallbackMap[topic] = null;
				delete _client.CallbackMap[topic];
			}
		}		
	}
		
	this.AddDisconnectHandler = function (handler){
		_client.disconnectHandlerMap.push(handler); 
	}
	
	this.RemoveDisconnectHandler = function (handler){
		var callbacksString = _client.disconnectHandlerMap.map(function (val, idx) {return '' + val;})
		var idx = $.inArray(''+handler, callbacksString);
		if (idx >= 0)
			_client.disconnectHandlerMap.splice(idx, 1);
	}
	
	this.SetAutoConnect = function (autoConnect) {
		_client.autoConnect = autoConnect;
		if (_client.autoConnect)
			_client.Connect();
	}
		
	this.Disconnect = function () {
		_client.autoConnect = false;
		_client.hubConnection.stop();
	}
	
	this.Connect = function () {
		_client.hubConnection.start().done(function () {
			_client.initialized = true;		
			for (var idx = 0; idx < _client.Queue.length; idx++)
			{
				var call = _client.Queue[idx];
				call[0].apply(null, call[1]);
			}
			
			var topics = Object.keys(_client.CallbackMap);
			for (var idx = 0; idx < topics.length; idx++)
			{
				if(_client.CallbackMap[topics[idx]])
					if (_client.CallbackMap[topics[idx]].length > 0)
						_client.proxy.invoke('subscribe', topics[idx]);
			}
		});
	}
	
	_client.proxy.on ('addMessage', function (topic, message, sender) {
		var callbacks = _client.CallbackMap[topic];
		if (callbacks != null)
		{
			for (var i = 0; i < callbacks.length; i++)
			{
				if (callbacks[i] != null)
					callbacks[i](topic, message, sender);
			}
		}			
	});
	
	_client.hubConnection.disconnected(function () {
		_client.initialized = false;
		_client.Queue = [];
		console.log('_client.hubConnection.disconnected:',_client, this,_client.disconnectHandlerMap.length);
		
		if(_client.disconnectHandlerMap)
			for (var i = 0; i < _client.disconnectHandlerMap.length; i++)
			{
				if (_client.disconnectHandlerMap[i] != null)
					_client.disconnectHandlerMap[i]();
			}
		if (_client.autoConnect)
			_client.Connect();
	});
	
	$(window).on('beforeunload', _client.Disconnect);
	this.Connect();
};
*/