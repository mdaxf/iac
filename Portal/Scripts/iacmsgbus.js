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