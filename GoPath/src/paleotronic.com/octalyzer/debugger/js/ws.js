var socket;
var shouldClose = false;
var lastEvent = new Date();

function handleWSMessage( event ) {
    console.log("Got Message")
    //console.log(event.data)

    lastEvent = new Date();
    var msg = JSON.parse(event.data)
    if (msg === null) {
        return;
    } 

    switch (msg['Type']) {
        case 'memsearch-response':
            console.log("memsearch-response "+JSON.stringify(msg))
            if (msg['Ok'] === true) {
                showSearchResult( msg['Payload'] );
            }
            break;  
        case 'live-rewind-response':
            console.log("live rewind "+JSON.stringify(msg))
            if (msg['Ok'] === true) {
                updateLRState( msg['Payload'] );
            }
            break;            
        case 'screen-update-response':
            //console.log("screen update "+JSON.stringify(msg))
            if (msg['Ok'] === true) {
                updateScreenshot();
            }
            break;
        case 'cpu-flag-response':
            if (msg['Ok'] === true) {
                updateCPUState( debugState.status, msg['Payload'] );
            } 
            break;
        case 'trace-response':
            if (msg['Ok'] === true) {
                message("success", "Trace", String(msg['Payload']));
            } 
            break;
        case 'heatmap-response':
            if (msg['Ok'] === true) {
                console.log(JSON.stringify(msg['Payload']))
                updateHeatmap(msg['Payload']);
            }
            break;  
        case 'get-settings-response':
            if (msg['Ok'] === true) {
                updateSettings(msg['Payload']);
            }
            break;           
        case 'sendkey-response':
            if (msg['Ok'] === true) {
                updateScreenshot();
            }
            break;
        case 'debug-message':
                if (msg['Ok'] === true) {
                    message('warning', 'Debug Message', String(msg['Payload']));
                }
                break;
        case 'setbp-response':
                if (msg['Ok'] === false) {
                    message('alert', 'Breakpoint', String(msg['Payload']));
                }
                break;
        case 'break-counter-response':
            if (msg['Ok'] === true) {
                updateCounters(msg['Payload']);
            }
            break;            
        case 'setpc-response':
            if (msg['Ok'] === true) {
                SendCommand("continue", ["now"]);
            }
            break;
        case 'attach-response':
            console.log("Handling attach response")
            if (msg['Ok'] === true) {
                message("success", "Attach success", "Attached to VM "+msg["Payload"]);
                $('#vm-number').html("VM #"+(msg['Payload']+1));
                SendCommand('get-settings', []);
                SendCommand("pause", ["now"]);
            } else {
                message("error", "Failed to attach to VM", "Failed to attach...");
            }    
            break;
        case 'break-response':
            console.log("Handling break response");
            if (msg['Ok'] === true) { 
                updateCPUStatus( "paused" );
                message('success', 'Paused', String(msg['Payload']));
                //if ($('#panel-stack').hasClass('is-active')) {
                    SendCommand( "getstack", [String(256)] );
                //    return;
                //}
            }
            break;
        case 'pause-response':
            console.log("Handling pause response: "+JSON.stringify(msg));
            if (msg['Ok'] === true) { 
                updateCPUState( "paused", msg['Payload'] );
                updateCPUStatus( "paused" );
                console.log("CPU state from pause response is "+debugState.status)
                message('success', 'Paused', 'CPU paused at '+sprintf("$%04X", debugState.cpu['PC']))
                //if ($('#panel-stack').hasClass('is-active')) {
                    SendCommand( "getstack", [String(256)] );
                //    return;
               // }
                requestUpdateHeatmap();
            }
            break;
        case 'get-memlock-response':
            console.log("Handling memlock response");
            if (msg['Ok'] === true) { 
                    $("#mem-lock").prop("checked", (msg['Payload'] === true));
                    //message("warning", "Memlock", String(msg['Payload']));
            }
            break;
        case 'state-response':
            console.log("Handling state response");
            if (msg['Ok'] === true) { 
                updateCPUState( debugState.status, msg['Payload'] );
                //if ($('#panel-stack').hasClass('is-active')) {
                    SendCommand( "getstack", [String(256)] );
                //    return;
                //}
                requestUpdateHeatmap();
            }
            break;
        case 'getstack-response':
            console.log("Handling getstack response");
            if (msg['Ok'] === true) { 
                updateCPUStack( msg['Payload'] );
            }
            break;
        case 'step-response':
            console.log("Handling step response");
            if (msg['Ok'] === true) { 
                updateCPUState( "paused", msg['Payload'] );
                updateCPUStatus( "paused" );
                message('success', 'Step', 'Stepped 1 instruction.')
                SendCommand( "getstack", [String(256)] );
            }
            break;
        case 'continue-response':
            console.log("Handling continue response");
            if (msg['Ok'] === true) { 
                updateCPUStatus( "freerun" );
                message('success', 'Running', 'Continuing execution.')
            }
            break;
        case 'decode-dasm-response':
            console.log("Handling decode response");
            if (msg['Ok'] === true) { 
                decodeInstructions( msg['Payload'], debugState.cpu['PC'], "#dasm-instructions", true );
            }            
            break;
        case 'decode-response':
            console.log("Handling decode response");
            if (msg['Ok'] === true) { 
                decodeInstructions( msg['Payload'], debugState.cpu['PC'], "#decode-instructions" );
                publishMemory();
            }            
            break;
        case 'memory-response':
            console.log("Handling memory update");
            if (msg['Ok'] === true) { 
                decodeMemory( msg['Payload'] );
            } 
            break;
        case 'switch-video-response':
            //console.log("Handling switch video response "+msg['Payload']);          
            if (msg['Ok'] === true) { 
                decodeVideoSwitches( msg['Payload'] );
            } 
            break;
        case 'listbp-response':
            //console.log("Handling list breakpoints response "+msg['Payload']);          
            if (msg['Ok'] === true) { 
                debugState.breakpoints = msg['Payload']['Breakpoints'];
                decodeBreakpoints( msg['Payload'] );
            } 
            break;            
    }

}

var WSConnected = false;

function InitWebsocket() {
    socket = new WebSocket("ws://localhost:9502/api/websocket/debug")
    socket.onmessage = handleWSMessage
    socket.onopen = function(ev) {
        WSConnected = true;
        console.log("Open "+ev.data)
        console.log("Websocket connected..." + ev.data)
        SendCommand( "attach", ["0"] );
    };
    socket.onerror = function(event) {
        console.log("Error "+event)
        WSConnected = false;
    };
    socket.onclose = function(event) {
        console.log("Close "+event)
        SendCommand("continue", []);
        WSConnected = false;
    };
    // This acts as a keep alive on the socket, and detects if the 
    // backend has gone away.
    setInterval( function() {
        var d = new Date();
        if (d.getTime() - lastEvent.getTime() > 5000) {
            SendCommand("ping", []);
        }
    }, 5000);
}

InitWebsocket();

function SendCommand( kind, args ) {
    socket.send( JSON.stringify( { Type: kind, Args: args } ) )
}
