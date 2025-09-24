var debugState = {
    cpu: {
        PC: 0,
        A: 0,
        X: 0,
        Y: 0,
        SP: 0,
        P: 0,
        Ahead: []
    },
    status: "",
    watch: {
        forceAux: 0,
        address: 0x400,
        size: 0x100,
        highlight: -1,
        highlightLen: 0,
    },
    cpulines: 0,
    screenms: 1000,
    screentimer: -1,
    breakpoints: []
};

//     Bit No.       7   6   5   4   3   2   1   0
//                   N   V       B   D   I   Z   C
const flagN = 128
const flagV = 64
const flagU = 32
const flagB = 16
const flagD = 8
const flagI = 4 
const flagZ = 2
const flagC = 1

var CaptureKeyboard = false;
var CaptureCPUControl = true;

function sendws() {
    var command = $('#command').val();
    var args = command.split(" ");
    var cmd = args[0];
    args.shift();
    SendCommand( 
        cmd,
        args,
    )
}

function registerPromptA() {
    registerPrompt("Enter new A register value", "6502.a", sprintf( "$%02x", debugState.cpu['A'] ), "A valid 8bit number");
}

function registerPromptX() {
    registerPrompt("Enter new X register value", "6502.x", sprintf( "$%02x", debugState.cpu['X'] ), "A valid 8bit number");
}

function registerPromptY() {
    registerPrompt("Enter new Y register value", "6502.y", sprintf( "$%02x", debugState.cpu['Y'] ), "A valid 8bit number");
}

function registerPromptPC() {
    registerPrompt("Enter new PC", "6502.pc", sprintf( "$%04x", debugState.cpu['PC'] ), "A valid 16bit number");
}

function registerPromptSP() {
    registerPrompt("Enter new SP", "6502.sp", sprintf( "$%04x", debugState.cpu['SP'] ), "A valid 16bit number");
}

function registerPrompt( prompt, field, value, error ) {
    $('#value-entry-prompt').html(prompt);
    $('#value-entry-target').val(field);
    $('#value-entry-value').val(value);
    $('#value-entry-message').html(error);
    $('#value-entry-dropdown').foundation('open');
}

function setRegister( reg, value ) {
    SendCommand( "set-register", [ String(reg), String(value) ] );
}

function toggleFlag( flag ) {
    console.log("Toggle flag "+flag);
    SendCommand("toggle-flag", [String(flag)]);
}

function flagString(flags) {
    var out = "";
    if (flags&flagN) {
        out += "<span onclick='toggleFlag(\"N\")' class='primary badge'>N</span>"
    } else {
        out += "<span onclick='toggleFlag(\"N\")'  class='secondary badge'>N</span>"
    }
    if (flags&flagV) {
        out += "<span onclick='toggleFlag(\"V\")'  class='primary badge'>V</span>"
    } else {
        out += "<span onclick='toggleFlag(\"V\")'  class='secondary badge'>V</span>"
    }
    if (flags&flagU) {
        out += "<span class='primary badge'>R</span>"
    } else {
        out += "<span class='secondary badge'>R</span>"
    }
    if (flags&flagB) {
        out += "<span onclick='toggleFlag(\"B\")' class='primary badge'>B</span>"
    } else {
        out += "<span onclick='toggleFlag(\"B\")' class='secondary badge'>B</span>"
    }
    if (flags&flagD) {
        out += "<span onclick='toggleFlag(\"D\")'  class='primary badge'>D</span>"
    } else {
        out += "<span onclick='toggleFlag(\"D\")'  class='secondary badge'>D</span>"
    }
    if (flags&flagI) {
        out += "<span onclick='toggleFlag(\"I\")'  class='primary badge'>I</span>"
    } else {
        out += "<span onclick='toggleFlag(\"I\")'  class='secondary badge'>I</span>"
    }
    if (flags&flagZ) {
        out += "<span onclick='toggleFlag(\"Z\")'  class='primary badge'>Z</span>"
    } else {
        out += "<span onclick='toggleFlag(\"Z\")'  class='secondary badge'>Z</span>"
    }
    if (flags&flagC) {
        out += "<span onclick='toggleFlag(\"C\")'  class='primary badge'>C</span>"
    } else {
        out += "<span onclick='toggleFlag(\"C\")'  class='secondary badge'>C</span>"
    }
    return out
}

// function updateScreenMS( ms ) {
//     debugState.screenms = ms;
//     if (debugState.screentimer !== -1) {
//         clearInterval(debugState.screentimer);
//     }
//     if (ms === 0) {
//         return;
//     }
//     debugState.screentimer = setInterval( function() {
//         updateScreenshot();
//     }, debugState.screenms );
// }

function getUrlParameter(sParam) {
    var sPageURL = decodeURIComponent(window.location.search.substring(1)),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : sParameterName[1];
        }
    }
};

function setUpdateFrequencyMS(ms) {
  SendCommand('setupdatems', [ String(ms) ]);
}

function attach(vm) {
    console.log("attach request for vm "+vm)
    if (vm < 0) {
        vm = 0;
    }
    SendCommand("attach", [String(vm)])
}

function detach() {
    SendCommand("detach", ["1"]);
    setTimeout(function() {
        window.close();
    }, 1000);
}

function getDASM(address) {
    var count = 20;

    console.log("Requesting "+count+" cpu decode lines from "+address);

    SendCommand( "decode-dasm", [String(address), String(count)] );
}

function getInstructions(address) {
    var count = debugState.cpulines;

    console.log("Requesting "+count+" cpu decode lines");

    SendCommand( "decode", [String(address), String(count)] );
}

function getStatus() {
    fetch('/api/debug/status')
    .then(function(resp) {
      console.log(resp);
      if (resp.ok) {
          resp.json().then(function(state) {
            console.log('Got cpu data: '+state);
            debugState.status = state.State
            $('#cpu-status').html(debugState.status)
          });
      } else {
        debugState.status = "disconnected"
        message( "error", "Failed to get CPU status", "Blah" );
      }
    });    
}

function stepCPU() {
    SendCommand("step", [""]);
}

function stepCPUOver() {
    SendCommand("step-over", [""]);
}

function stepCPUOut() {
    SendCommand("continue-out", [""]);
}

function runCPU() {
    SendCommand("continue", [""]);
}

function pauseCPU() {
    SendCommand("pause", [""]);
}

function updateCPUStack( stack ) {

    //
    console.log("Stack: "+JSON.stringify(stack));
    var out = "";
    var addr = 0x200 - stack.Values.length;
    for (v of stack.Values) {
        var c = " ";
        if (v > 32 && v < 127) {
            c = String.fromCharCode(v);
        }
        out += sprintf("<tr><td class='register'>$%03x:</td><td class='register'><a href='javascript:editAddress(%d, %d)'>%02x</a></td><td class='register'>%d</td><td class='register'>%s</td></tr>", addr, addr, Number(v), Number(v), Number(v), c);
        addr++;
    }

    $('#decode-stack').html(out);
}

function updateCPUStatus( status ) {
    debugState.status = status
    $('#cpu-status').html(status);

    switch (status) {
    case "freerun":
        $("#cpu-step").prop('disabled', true);
        $("#cpu-stepo").prop('disabled', true);
        $("#cpu-stepx").prop('disabled', true);
        $("#cpu-pause").prop('disabled', false);
        $("#cpu-continue").prop('disabled', true);
        $("#cpu-go").prop('disabled', false);
        $("#cpu-trace").prop('disabled', false);
        $("#cpu-trace-off").prop('disabled', false);
        $("#cpu-speed-025x").prop('disabled', false);
        $("#cpu-speed-05x").prop('disabled', false);
        $("#cpu-speed-1x").prop('disabled', false);
        $("#cpu-speed-2x").prop('disabled', false);
        $("#cpu-speed-4x").prop('disabled', false);
        break;
    case "paused":
        $("#cpu-step").prop('disabled', false);
        $("#cpu-stepo").prop('disabled', false);
        $("#cpu-stepx").prop('disabled', false);
        $("#cpu-pause").prop('disabled', true);
        $("#cpu-continue").prop('disabled', false);
        $("#cpu-go").prop('disabled', false);
        $("#cpu-trace").prop('disabled', false);
        $("#cpu-trace-off").prop('disabled', false);
        $("#cpu-speed-025x").prop('disabled', true);
        $("#cpu-speed-05x").prop('disabled', true);
        $("#cpu-speed-1x").prop('disabled', true);
        $("#cpu-speed-2x").prop('disabled', true);
        $("#cpu-speed-4x").prop('disabled', true);
        break;
    }

    console.log("===== CPU IS "+debugState.status+" =====")
}

function updateCPUState( mode, state ) {
    debugState.status = mode
    debugState.cpu = state
    $('#cpu-pc').html( sprintf( "<a href='javascript:void' onclick='registerPromptPC();'>$%04x</a>", debugState.cpu['PC']));
    $('#cpu-a').html(sprintf( "<a href='javascript:void' onclick='registerPromptA();'>$%02x</a>", debugState.cpu['A']));
    $('#cpu-x').html(sprintf( "<a href='javascript:void' onclick='registerPromptX();'>$%02x</a>", debugState.cpu['X']));
    $('#cpu-y').html(sprintf( "<a href='javascript:void' onclick='registerPromptY();'>$%02x</a>", debugState.cpu['Y']));
    $('#cpu-sp').html(sprintf( "<a href='javascript:void' onclick='registerPromptSP();'>$%04x</a>", debugState.cpu['SP']));
    $('#cpu-p').html(flagString(debugState.cpu['P']));
    if (debugState.cpu['Speed'] != null) {
        switch (Math.round(100 * Number(debugState.cpu['Speed']))) {
        case 25:
           $('#cpu-speed-025x').prop('disabled', true);
           $('#cpu-speed-05x').prop('disabled', false);
           $('#cpu-speed-1x').prop('disabled', false);
           $('#cpu-speed-2x').prop('disabled', false);
           $('#cpu-speed-4x').prop('disabled', false);
           break; 
        case 50:
            $('#cpu-speed-025x').prop('disabled', false);
            $('#cpu-speed-05x').prop('disabled', true);
            $('#cpu-speed-1x').prop('disabled', false);
            $('#cpu-speed-2x').prop('disabled', false);
            $('#cpu-speed-4x').prop('disabled', false);
            break;
        case 100:
            $('#cpu-speed-025x').prop('disabled', false);
            $('#cpu-speed-05x').prop('disabled', false);
            $('#cpu-speed-1x').prop('disabled', true);
            $('#cpu-speed-2x').prop('disabled', false);
            $('#cpu-speed-4x').prop('disabled', false);
            break; 
        case 200:
            $('#cpu-speed-025x').prop('disabled', false);
            $('#cpu-speed-05x').prop('disabled', false);
            $('#cpu-speed-1x').prop('disabled', false);
            $('#cpu-speed-2x').prop('disabled', true);
            $('#cpu-speed-4x').prop('disabled', false);
            break;         
        case 400:
            $('#cpu-speed-025x').prop('disabled', false);
            $('#cpu-speed-05x').prop('disabled', false);
            $('#cpu-speed-1x').prop('disabled', false);
            $('#cpu-speed-2x').prop('disabled', false);
            $('#cpu-speed-4x').prop('disabled', true);
            break; 
        }
        //$('#cpu-speed-label').html( sprintf( "%.02f", debugState.cpu['Speed'] ) + "x" );
    }
    if (debugState.status === "paused" || debugState.status === "halted") {
        getInstructions( debugState.cpu['PC'] );
        //updateScreenshot();
    } //else {
    //    $('#decode-instructions').html("Pause / Step CPU to see instructions.")
    //}
    $('#cpu-cc').html(debugState.cpu['CC']);
    refreshBreakpoints(); // only happens if tab active
    refreshMessages();
}

function refreshMessages() {
    SendCommand( "get-lastbpmsg", [""] );
}

function getRegisters() {
    fetch('/api/debug/cpu')
    .then(function(resp) {
      console.log(resp);
      if (resp.ok) {
          resp.json().then(function(state) {
            console.log('Got cpu data: '+state);
            debugState.cpu = state
            $('#cpu-pc').html('$'+sprintf( "%04x", debugState.cpu['PC']));
            $('#cpu-a').html('$'+sprintf( "%02x", debugState.cpu['A']));
            $('#cpu-x').html('$'+sprintf( "%02x", debugState.cpu['X']));
            $('#cpu-y').html('$'+sprintf( "%02x", debugState.cpu['Y']));
            $('#cpu-sp').html('$'+sprintf( "%03x", debugState.cpu['SP']));
            $('#cpu-p').html(flagString(debugState.cpu['P']));
            if (debugState.status === "paused" || debugState.status === "halted") {
                getInstructions( debugState.cpu['PC'] );
            }
            //     $('#decode-instructions').html("Pause / Step CPU to see instructions.")
            // }
          });
      } else {
        debugState.cpu = {
            PC: 0,
            A: 0,
            X: 0,
            Y: 0,
            SP: 0,
            P: 0,
            Ahead: []
        }
        message( "error", "Failed to get CPU state", "Blah" );
      }
    });
}

function traceOn() {
    SendCommand("trace", ["on"]);
}

function traceOff() {
    SendCommand("trace", ["off"]);
}

function decodeVideoSwitches( rec ) {
    var out = ""
    for (inst of rec.Switches) {
        var cls = 'register';
        var txt = '';
        if (inst['Enabled']) {
            cls += " switchon";
            txt = inst['EnabledText'];
        } else {
            cls += " switchoff";
            txt = inst['DisabledText'];
        }
        out += "<tr class='register'>"
        out += "<td class='register'>"
        out += inst['Name']
        out += "</td>"
        out += "<td class='register'>"+sprintf("$%02x", inst['StatusAddress']&0xFF)+"</td>"
        out += "<td align='center' valign='center' class='register table-20'><button class='"+cls+"' onclick='toggleSwitch(\""+inst['Name']+"\")'>"+txt+"</button></td>"
        out += "</tr>"
    };
    //$('#decode-video-switches-1').html(out)
    $('#decode-video-switches-2').html(out)
}

var firstInstSize = 1;

function decodeInstructions( instructions, address, target, hotlink ) {

    if (hotlink == null) {
        hotlink = false;
    }

    if (target == null || target === "") {
        target = "#decode-instructions";
    } else {
        firstInstSize = instructions.Instructions[0].Bytes.length;
    }

    var out = ""
    var cc = "";
    for (inst of instructions.Instructions) {
        //console.log("Compare "+inst.Address+" and "+address);
        if (String(inst.Address) === String(address) && inst.Historic !== true) {
            cc = ' current';
        } else {
            cc = ''
        }
        out += "<tr>"
        out += "<td class='register"+cc+"'>"
        if (hotlink) {
            out += sprintf( "<a href='javascript:void();' onclick='pcMenu(0x%04x)'>$%04x</a>", inst.Address, inst.Address );
        } else {
            out += sprintf( "$%04x", inst.Address );
        }
        out += "</td>"
        out += "<td class='register"+cc+"'>"
        for (b of inst.Bytes) {
            out += sprintf("%02x ", b)
        }
        out += "</td>"
        out += "<td class='register"+cc+"'>"
        out += inst.Instruction
        out += "</td>"
        out += "<td class='register"+cc+"'>"
        out += inst.Cycles
        out += "</td>"
        out += "</tr>"

    };
    $(target).html(out)
    //updateScreenshot();
}

function removeBp(index) {
    SendCommand("clrbp", [String(index)]);
}

function disableBp(index) {
    SendCommand("disbp", [String(index)]);
}

function enableBp(index) {
    SendCommand("enbp", [String(index)]);
}

function bpInfo(bp) {

    console.log("bpInfo("+JSON.stringify(bp)+")")

    var out = ""
    if (bp["ValuePC"] !== null) {
        out += sprintf("PC=$%04x ", bp["ValuePC"]);
    } 
    if (bp["MaxCount"] !== null) {
        out += sprintf("MAXCNT=%d ", bp["MaxCount"]);
    } 
    if (bp["ValueA"] !== null) {
        out += sprintf("A=$%02x ", bp["ValueA"]);
    } 
    if (bp["ValueX"] !== null) {
        out += sprintf("X=$%02x ", bp["ValueX"]);
    } 
    if (bp["ValueY"] !== null) {
        out += sprintf("Y=$%02x ", bp["ValueY"]);
    } 
    if (bp["ValueSP"] !== null) {
        out += sprintf("SP=$%04x ", bp["ValueSP"]);
    } 
    if (bp["ValueP"] !== null) {
        out += sprintf("P=$%02x ", bp["ValueP"]);
    } 
    if (bp["WriteAddress"] !== null) {
        out += sprintf("WA=$%04x ", bp["WriteAddress"]);
    } 
    if (bp["WriteValue"] !== null) {
        out += sprintf("WV=$%02x ", bp["WriteValue"]);
    } 
    if (bp["ReadAddress"] !== null) {
        out += sprintf("RA=$%04x ", bp["ReadAddress"]);
    } 
    if (bp["ReadValue"] !== null) {
        out += sprintf("RV=$%02x ", bp["ReadValue"]);
    } 
    console.log("main "+bp['Main']);
    console.log("aux "+bp['Aux']);
    if (bp['Main'] === 1) {
        out += sprintf("MAIN=%d ", bp['Main']);
    }
    if (bp['Aux'] === 1) {
        out += sprintf("AUX=%d ", bp['Aux']);
    }
    if (bp["Ephemeral"] != null && bp["Ephemeral"]) {
        out += sprintf("EPH=%d ", 1);
    } 

    if (bp["Action"] != null) {
        out += sprintf("ACTID=%s ", bp["Action"]["Type"]);
        switch (bp["Action"]["Type"]) {
        case 0:
            out += "(break)";
            break;    
        case 1:
            out += sprintf("(message: %s)", bp["Action"]["Arg0"].replace(/_/g, " "));
            break;
        case 2:
            out += "(chime)";
            break;
        case 3:
            out += "(start trace)";
            break;
        case 4:
            out += "(stop trace)";
            break;
        case 5:
            out += sprintf("(jump to $%04x)", bp["Action"]["Arg0"]);
            break;
        case 6:
            out += "(inc counter)";
            break;
        case 7:
            out += sprintf("(cpu speed: %f)", bp["Action"]["Arg0"]);
            break;
        case 8:
            out += sprintf("(log trace: %s)", bp["Action"]["Arg0"].replace(/_/g, " "));
            break;
        case 9:
            out += "(start recording)";
            break;
        case 10:
            out += "(stop recording)";
            break;
        }
        // if (bp["Action"]["Type"] == "1" || bp["Action"]["Type"] == "8") {
        //     out += sprintf("ACTTXT=%s ", bp["Action"]["Arg0"]);
        // } else if (bp["Action"]["Type"] == "5") {
        //     out += sprintf("ACTJMP=$%04x ", bp["Action"]["Arg0"]);
        // } else if (bp["Action"]["Type"] == "7") {
        //     out += sprintf("ACTSPD=%f ", bp["Action"]["Arg0"]); 
        // }
    } 

    return out
}

function decodeBreakpoints( breakpoints ) {
    var out = ""
    var count = 1;
    for (bp of breakpoints.Breakpoints) {
        if (bp['Disabled'] === true) {
            out += "<tr class='disabled-bp'>"
        } else {
            out += "<tr>"  
        }
        out += "<td>"
        out += sprintf( "%02x", count )
        out += "</td>"
        out += "<td>"  
        out += bpInfo(bp)
        out += "</td>"
        out += "<td>"
        if (bp['Counter'] != null && bp['Counter'] != 0) {
            out += "<a href='javascript:void;' onclick='clearCounter("+count+")'>"+bp['Counter']+"</a>";
        }
        out += "</td>"

        out += "<td style='text-align: right;'>"
        out += "<a href='javascript:void();' class='hollow tiny' onclick='removeBp("+count+");'>"+icon("delete", "red")+"</button>"   

        out += "&nbsp;"

        if (bp['Disabled'] === true) {
            out += "<a href='javascript:void();' class='hollow tiny' onclick='enableBp("+count+");'>"+icon("unchecked", "#ff0000")+"</button>"     
        } else {
            out += "<a href='javascript:void();' class='hollow tiny' onclick='disableBp("+count+");'>"+icon("checked", "orange")+"</button>"             
        }

        out += "&nbsp;"

        out += "<a href='javascript:void();' class='hollow tiny' onclick='editBp("+count+");'>"+icon("edit")+"</a>"

        out += "</td>"
        out += "</tr>"
        count++;
    };
    $('#breakpoint-list').html(out)
}

function editAddress( address, value ) {
    $('#mem-address').val(sprintf("$%04x", address));
    $('#mem-value').val(sprintf("$%02x", value));
    var c = value;
    if (c < 0x20 || c > 0x7f) {
        c = 0x20
    }
    if (c !== 0x20) {
        $('#mem-ascii').val(sprintf("%s", String.fromCharCode(c)));
    } else {
        $('#mem-ascii').val("");
    }
    SendCommand("get-memlock", [ String(address) ]);
    $('#mem-edit-dropdown').foundation('open');
}

function updateASCIICode() {
    var c = $('#mem-ascii').val();
    if (c === "") {
        c = " ";
    }
    var s = sprintf( "$%02x", c.charCodeAt(0) );
    $('#mem-value').val(s);
}

function asmSubmit() {
    // TODO: submit asm request
    var addr = $('#asm-address').val();
    if (addr == null || addr == "") {
        $('#asm-status').html("Error: specify address.");
        return;
    }
    $('#asm-status').html("Submitted.");
}

function pcMenu(addr) {
    $('#pc-menu-addr').val( sprintf("$%04x", addr) );
    $('#pc-menu-mode option:eq("0")').prop('selected', true);
    $('#pc-menu-dropdown').foundation('open');
}

function pcMenuSubmit() {
    // do something here
    var mode = $('#pc-menu-mode').val();
    var addr = $('#pc-menu-addr').val();
    switch (Number(mode)) {
    case 0:
        SendCommand("setbp", [ sprintf("PC=%s", addr), "EPH=1" ]);
        break
    case 1:
        SendCommand("setbp", [ sprintf("PC=%s", addr) ]);
        break
    case 2:
        SendCommand("setpc", [ sprintf("$%04x", Number(addr) ) ] );
        break
    }
}

function updateSelectBank( bank ) {
    var mode = $('#heatmap-mode').val();
    if (mode === 1 || mode === 2) {
        mode += 2;
        console.log("HEATMAP MODE = "+mode);
        console.log("HEATMAP BANK = "+bank);
        $('#heatmap-bank').val(bank);
        // $('#heatmap-mode').val(mode);
        $('#heatmap-mode option:eq('+mode+')').prop('selected', true);
        updateHeatmapMode();
    }
}

function reset() {
    SendCommand( "apple-reset-6502", [""] );
}

function updateHeatmapMode() {
    var mode = $('#heatmap-mode').val();
    var bank = $('#heatmap-bank').val();
    if (bank == null || bank == "") {
        bank = 0;
    }
    console.log("heatmap mode is "+mode);
    console.log("heatmap bank is "+bank);
    SendCommand("set-heatmap", [ String(mode), String(bank) ]);
}

function requestUpdateHeatmap() {
    if (!$('#panel-memory').hasClass('is-active')) {
        return;
    }
    if (!$('#panel-heatmap').hasClass('is-active')) {
        return;
    }
    SendCommand("get-heatmap", [""]);
}

function updateHeatmap( heatmap ) {
    // update heatmap content if active
    if (!$('#panel-memory').hasClass('is-active')) {
        return;
    }

    if (!$('#panel-heatmap').hasClass('is-active')) {
        return;
    }

    var mode = $('#heatmap-mode').val();
    var bankClick = (mode === 1 || mode === 2);

    var out = "<table>";
    out += "<thead>";
    out += "<tr class='register'>";
    out += "<th class='register'></th>";
    out += "<th class='register'>x0</th>";
    out += "<th class='register'>x1</th>";
    out += "<th class='register'>x2</th>";
    out += "<th class='register'>x3</th>";
    out += "<th class='register'>x4</th>";
    out += "<th class='register'>x5</th>";
    out += "<th class='register'>x6</th>";
    out += "<th class='register'>x7</th>";
    out += "<th class='register'>x8</th>";
    out += "<th class='register'>x9</th>";
    out += "<th class='register'>xA</th>";
    out += "<th class='register'>xB</th>";
    out += "<th class='register'>xC</th>";
    out += "<th class='register'>xD</th>";
    out += "<th class='register'>xE</th>";
    out += "<th class='register'>xF</th>";
    out += "</tr>"
    out += "</thead>";
    out += "<tbody>";

    var v = 0;
    var wr = 0;
    var rd = 0;
    var col = "ffffff";
    var rr, gg, bb = 0;
    var bank = "";

    console.log("mode is "+mode);

    for (r = 0; r<16; r++) {
        out += "<tr class='register'><th class='register'>" + sprintf( "%x", r ) + "x</th>";
        for (c = 0; c<16; c++) {
            v = heatmap[r*16+c];

            rr = 0xff
            gg = 0xff
            bb = 0xff

            if (v != 0) {
                wr = (v & 0xf0) | 0xf;
                rd = ((v & 0x0f) << 4) | 0xf;

                if (mode === "5") {
                    bb -= (wr+rd)/2;
                    rr -= (wr/2);
                    gg -= (rd/2);
                } else {
                    gg -= (wr+rd)/2;
                    rr -= (rd/2);
                    bb -= (wr/2);
                }
            }

            col =  sprintf("%02x%02x%02x", rr,gg,bb);
            out += "<td class='register' style='background: #" + col + "'>";
            //if (bankClick) {
                bank = sprintf("%02x", r*16+c)
                // out += "<a href='javascript:updateSelectBank("+(r*16+c)+");'>"+bank+"</a>";
            //} else {
            out += "<span class='bank'>&nbsp;&nbsp;"+bank+"</span>";
            //}
            out += "</td>";  
        }
        out += "</tr>";
    }

    out += "</tbody>";
    out += "</table>";

    $('#heatmap-value').html(out);
}

function updateScreenshot() {
    if (!$('#panel-hgr').hasClass('is-active')) {
        return;
    }
    var img = document.getElementById("screenshot");
    var tag = Date.now();
    img.src = sprintf("/api/debug/screen/?t=%d", tag);
    console.log("Load image");
}

function poke2ascii(v) {

    v = v & 0xff

	if (v >= 0 && v <= 31) {
		return (64 + (v % 32))
	}

	if (v >= 32 && v <= 63) {
		return (32 + (v % 32))
	}

	if (v >= 64 && v <= 95) {
		return (64 + (v % 32))
	}

	if (v >= 96 && v <= 127) {
		return (32 + (v % 32)) 
	}

	if (v >= 128 && v <= 159) {
		return (64 + (v % 32))
	}

	if (v >= 160 && v <= 191) {
		return (32 + (v % 32))
	}

	if (v >= 192 && v <= 223) {
		return (64 + (v % 32))
	}

	if (v >= 224 && v <= 255) {
		return 96 + (v % 32)
    }
    
    return v
}

function decodeMemory( memory ) {
    var hexstr = ""
    var asciistr = ""
    var out = ""
    var count = 0

    var stripHighBit = $('#mem-ascii-strip').is(":checked");

    var w = $('#memory-view').width();
    console.log("WIDTH = "+w)
    var perline = 16;
    // if (w < 700) {
    //     perline = 12;
    // }

    out += "<table><thead><tr class='register'><th>Memory</th><th>ASCII</th></tr></thead><tbody>";

    for (data of memory.Memory) {

        var useclass = "register hexbutton";
        if (debugState.watch.highlight !== -1 && memory.Address+count >= debugState.watch.highlight && memory.Address+count < debugState.watch.highlight+debugState.watch.highlightLen) {
            useclass = "register hexbutton highlight";
        }

        var refAddr = (memory.Address + count) % 65536

        if (count % perline == 0) {
            if (hexstr.length > 0) {
                out += hexstr + "</td><td class='register'>" + asciistr + "</td></tr>" 
            }
            hexstr = sprintf( "<tr class='register'><td class='register'>%04x: ", refAddr )
            asciistr = ""
        }
        var displayval = Number(data);
        hexstr += sprintf(" <a class='"+useclass+"' href='javascript:editAddress(%d, %d);'>%02x</a>", refAddr, data, data);
        if (stripHighBit === true) {
            displayval = poke2ascii(displayval);
        }
        asciistr += sprintf("<a class='register hexbutton' href='javascript:editAddress(%d, %d);'>", refAddr, data );
        if ((displayval >= 32) && (displayval <= 127)) {
            asciistr += String.fromCharCode(displayval);
        } else {
            asciistr += "."
        }
        asciistr += "</a>"
        count++;
    };
    if (hexstr.length > 0) {
        out += hexstr + "</td><td class='register'>" + asciistr + "</td></tr>"  
    }
    out += "</tbody></table>";
    $('#decode-memory').html(out)
}

function publishMemory() {
    if (!$('#panel-memory-edit').hasClass('is-active')) {
        return;
    }

    var forceAux = 0;
    if ($('#force-aux').is(":checked")) {
        forceAux = 1
    }
    SendCommand("memory", [String(debugState.watch.address), String(debugState.watch.size), String(forceAux)]);
}

function saveState() {
    SendCommand("pause", [])
    SendCommand("state-save", []);
}

function toggleSwitch(label) {
    console.log("Toggle softswitch "+label)
    SendCommand("toggle-switch", [label]);
}

function showSearchResult( result ) {
    var addr = Number(result["FoundAddr"]);
    var count = result["Search"].length;
    debugState.watch.address = addr;
    debugState.watch.highlight = addr;
    debugState.watch.highlightLen = count;
    if (result["Aux"] == true) {
        debugState.watch.forceAux = 1;
    } else {
        debugState.watch.forceAux = 0;
    }
    publishMemory();
}

function searchMemory() {
    var forceAux = 0;
    if ($('#force-aux').is(":checked")) {
        forceAux = 1
    }
    var highBit = $('#mem-search-high').is(":checked");
    var address = debugState.watch.address;
    var value   = $('#mem-search-value').val();
    var values = value.split( " " );

    if (value.length === 0) {
        value = $('#mem-search-string').val();
        if (value.length > 0) {
            values = [];
            for (var i=0; i<value.length; i++) {
                if (highBit === true) {
                    values.push( sprintf("$%02x", value.charCodeAt(i) | 128) );
                } else {
                    values.push( sprintf("$%02x", value.charCodeAt(i)) );
                }
            }
            console.log("Input was "+value);
            console.log("String Values are "+JSON.stringify(values))
        }
    }

    debugState.watch.highlight = -1; // clear 

    values.unshift( String(forceAux) );
    values.unshift( String(address) );
    console.log("Request memory search with params: "+JSON.stringify(values));
    SendCommand( "memsearch", values );
}

function updateMemory() {
    var forceAux = 0;
    if ($('#force-aux').is(":checked")) {
        forceAux = 1
    }
    var address = $('#mem-address').val();
    var value   = $('#mem-value').val();
    var locked  = $('#mem-lock').is(":checked", true) ? 1 : 0;
    console.log("updating memory address %s to %s", address, value)
    SendCommand("memset", [String(address), String(value), String(forceAux), String(locked)]);
    setTimeout( function() {
        publishMemory();
        SendCommand( "getstack", ["256"] );
    }, 250);
}

function memAddrPC() {
    debugState.watch.address = Number(debugState.cpu.PC) & 0xff00;
    publishMemory();
}

function nextPage(step) {
    debugState.watch.address = Number(debugState.watch.address) + Number(step)*256;
    if (debugState.watch.address >= 65536) {
        debugState.watch.address -= 65536
    }
    publishMemory();
}

function prevPage(step) {
    debugState.watch.address = Number(debugState.watch.address) - Number(step)*256;
    if (debugState.watch.address < 0) {
        debugState.watch.address += 65536
    }
    publishMemory();
}

function clearAllBreakpoints() {
    SendCommand( "clrbp", ["*"] );
}

function updateSettings(p) {
    console.log(JSON.stringify(p))
    $('#cpu-update-ms').val( p['CPUStateInterval'] );
    $('#cpu-backlog-lines').val( p['CPUInstructionBacklog'] );
    $('#cpu-lookahead-lines').val( p['CPUInstructionLookahead'] );
    $('#screen-refresh-ms').val( p['ScreenRefreshMS'] );
    $('#cpu-record-timing').val( p['CPURecordTiming'] );
    debugState.cpulines = Number(p['CPUInstructionBacklog']) + Number(p['CPUInstructionLookahead']);
    console.log("Set cpu lines to "+debugState.cpulines);
    $('#cpu-full-record').prop('checked', p['FullCPURecord']);
    $('#cpu-break-brk').prop('checked', p['BreakOnBRK']);
    $('#cpu-break-ill').prop('checked', p['BreakOnIllegalOp']);
    $('#save-on-submit').prop('checked', p['SaveSettingsOnSubmit']);
    message("success", "Settings", "Updated.");
}

function publishSettings() {
    message("warning", "Settings", "Publishing.");
    SendCommand( "set-val", [ "cpu-update-ms", String($('#cpu-update-ms').val()) ] );
    SendCommand( "set-val", [ "cpu-backlog-lines", String($('#cpu-backlog-lines').val()) ] );
    SendCommand( "set-val", [ "cpu-lookahead-lines", String($('#cpu-lookahead-lines').val()) ] );
    SendCommand( "set-val", [ "cpu-record-timing", String($('#cpu-record-timing').val()) ] );
    SendCommand( "set-val", [ "screen-refresh-ms", String($('#screen-refresh-ms').val()) ] );
    var fullrec = $('#cpu-full-record').is(':checked');
    if (fullrec === true) {
        SendCommand( "set-val", [ "cpu-full-record", "1" ] );
    } else {
        SendCommand( "set-val", [ "cpu-full-record", "0" ] )
    }
    var breakBRK = $('#cpu-break-brk').is(':checked');
    if (breakBRK === true) {
        SendCommand( "set-val", [ "cpu-break-brk", "1" ] );
    } else {
        SendCommand( "set-val", [ "cpu-break-brk", "0" ] )
    }
    var breakILL = $('#cpu-break-ill').is(':checked');
    if (breakILL === true) {
        SendCommand( "set-val", [ "cpu-break-ill", "1" ] );
    } else {
        SendCommand( "set-val", [ "cpu-break-ill", "0" ] )
    }
    var save = $('#save-on-submit').is(':checked');
    if (save === true) {
        SendCommand( "set-val", [ "save-on-submit", "1" ] );
    } else {
        SendCommand( "set-val", [ "save-on-submit", "0" ] )
    }
}

function openUploader() {
    $('#mem-load-dropdown').foundation('open');
}

function openASMUploader() {
    $('#asm-load-dropdown').foundation('open');
}

function openDownloader() {
    $('#mem-save-mode').val("bin");
    $('#mem-save-dropdown').foundation('open');
}

function openDasmDownloader() {
    $('#mem-save-mode').val("dasm");
    $('#mem-save-dropdown').foundation('open');
}

function openDumpDownloader() {
    $('#mem-save-mode').val("txt");
    $('#mem-save-dropdown').foundation('open');
}

function downloadBlob() {
    var mode = $('#mem-save-mode').val();
    var addr = $('#mem-save-address').val();
    var len = $('#mem-save-length').val();
    var url = sprintf("/api/debug/download/%s/%s/%s", mode, addr, len);
    var a = document.createElement("a");
    a.href = url;
    a.click();
}

function uploadBlob() {
    var f = $('#mem-load-file')[0].files[0];
    var addr = $('#mem-load-address').val();
    readFileIntoMemory( f, function(info) {
        // something with info...
        console.log("Got file of size "+info.size);
        const options = {
            method: 'post',
            headers: {
              'Content-type': 'application/octet-stream; charset=UTF-8',
              'X-LoadAddress': String(addr),
            },
            body: info.content,
          }
          
          fetch("/api/debug/upload", options).then(response => {
              console.log("Response: "+JSON.stringify(response))
          }).catch(err => {
            console.error('Request failed', err)
            return;
          })

          console.log("Upload succeeded!")
    });
}

function uploadASMBlob() {
    var f = $('#asm-load-file')[0].files[0];
    var addr = $('#asm-load-address').val();
    readFileIntoMemory( f, function(info) {
        // something with info...
        console.log("Got file of size "+info.size);
        const options = {
            method: 'post',
            headers: {
              'Content-type': 'application/octet-stream; charset=UTF-8',
              'X-LoadAddress': String(addr),
            },
            body: info.content,
          }
          
          fetch("/api/debug/asm/upload", options).then(response => {
              if (response.ok) {
                message( "success", "ASM Success", "ASM completed successfully. ");
              } else {
                message( "alert", "ASM Failed", "" );  
              }
              console.log("Response: "+JSON.stringify(response))
          }).catch(err => {
            message( "alert", "ASM Failed", err );
            console.error('Request failed', err);
            return;
          })

          console.log("Upload succeeded!")
    });
}

function readFileIntoMemory (file, callback) {
    var reader = new FileReader();
    reader.onload = function () {
        callback({
            name: file.name,
            size: file.size,
            type: file.type,
            content: new Uint8Array(this.result)
         });
    };
    reader.readAsArrayBuffer(file);
}

var dasmAddr = 0x2000;

function dasmAddrPC() {
    dasmAddr = debugState.cpu.PC;
    getDASM(dasmAddr);
}

function dasmForward() {
    console.log("moving forward "+firstInstSize+" instructions.")
    dasmAddrChange( firstInstSize );
}

function dasmAddrChange( n ) {
    console.log("dasmAddr = "+dasmAddr);
    console.log("modify dasm addr "+n);
    dasmAddr += Number(n);
    if (dasmAddr < 0) {
        dasmAddr = 0;
    }
    if (dasmAddr > 65535) {
        dasmAddr = 65535;
    }
    getDASM(dasmAddr);
}

function asNumber( n ) {
    n = String(n);
    if (n.startsWith("$")) {
        n = n.replace( "$", "0x" );
    }
    return parseInt(n);
}

function dasmCall() {
    var addr = $('#memory-dasm-address').val();
    dasmAddr = asNumber(addr);
    //$('#dasm-addr').html(sprintf("$%04x", dasmAddr));
    getDASM(dasmAddr);
}

function submitRegisterUpdate() {
    var field = $('#value-entry-target').val();
    var value = $('#value-entry-value').val();
    console.log('set '+field+' to '+value);
    SendCommand("set-val", [ String(field), String(value) ]);
    setTimeout( function () {

    }, 250 );
}

var active = true;

function updateBreakAction() {
    var action = $('#break-action').val();
    if (action == null || action == "") {
        action = 0;
    }
    $('#break-jump-addr').prop('disabled', !(action == "5"));
    $('#break-text').prop('disabled', !(action == "1" || action == "8"));
    $('#break-speed').prop('disabled', !(action == "7"));
}

function initialize() {
    debugState.cpu = {
        PC: 0,
        A: 0,
        X: 0,
        Y: 0,
        SP: 0,
        P: 0,
    }
    //attach(0);
   // getRegisters();
    //getStatus();
    //getInstructions(debugState.cpu['PC'], 20);

    window.onclose = function() {
        SendCommand("detach", ["1"]);
    };

    $('#mem-ascii-strip').change(function() {
        console.log("Toggle high bit");
        publishMemory();
    });

    $('textarea').keydown(function(e) {
        console.log("Event"+JSON.stringify(e))
        if (e.keyCode === 9) {
            console.log("Tab pressed");
            // get caret position/selection
            var start = this.selectionStart;
            var end = this.selectionEnd;
    
            var $this = $(this);
            var value = $this.val();
    
            // set textarea value to: text before caret + tab + text after caret
            $this.val(value.substring(0, start)
                        + "\t"
                        + value.substring(end));
    
            // put caret at right position again (add one for the tab)
            this.selectionStart = this.selectionEnd = start + 1;
    
            // prevent the focus lose
            e.preventDefault();
        }
    });

    $('#command').keypress(function(event) {
        console.log("Event"+JSON.stringify(event))
        if (event.keyCode === 13) {
            console.log("Enter pressed");
            $('#commandButton').click();
        }
    });
    
    $('#commandButton').click(function() {
        console.log("Button has been clicked");
        sendws();
    });

    $('#sttab').click(function() {
        SendCommand(
            "getstack",
            []
        );
    });


    $('#bptab').click(function() {
        SendCommand(
            "listbp",
            []
        );
    });

    $('#memtab').click(function() {
        publishMemory();
    });

    $('#hmtab').click(function() {
        SendCommand(
            "get-heatmap",
            []
        );
    });

    $('#settingstab').click(function() {
        SendCommand(
            "get-settings",
            []
        );
    });

    $('#asm-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        asmSubmit();
        //$('#pc-menu-dropdown').foundation('close');
        return false;
    });

    $('#pc-menu-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        pcMenuSubmit();
        $('#pc-menu-dropdown').foundation('close');
        return false;
    });

    $('#value-entry-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        submitRegisterUpdate();
        $('#value-entry-dropdown').foundation('close');
        return false;
    });

    $('#dasm-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        dasmCall();
        $('#dasm-dropdown').foundation('close');
        return false;
    });

    $('#add-bp-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        addBreakpoint();
        $('#add-bp-dropdown').foundation('close');
        return false;
    });

    $('#mem-edit-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        updateMemory();
        $('#mem-edit-dropdown').foundation('close');
        return false;
    });

    $('#asm-load-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        uploadASMBlob();
        $('#asm-load-dropdown').foundation('close');
        return false;
    });

    $('#mem-load-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        uploadBlob();
        $('#mem-load-dropdown').foundation('close');
        return false;
    });

    $('#mem-save-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        downloadBlob();
        $('#mem-save-dropdown').foundation('close');
        return false;
    });

    $('#mem-search-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        searchMemory();
        $('#mem-search-dropdown').foundation('close');
        return false;
    });

    $('#settings-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        publishSettings();
        //$('#settings-dropdown').foundation('close');
        return false;
    });

    $('#screenshot').mouseenter( function() {
        console.log("Mouse entered screen")
        CaptureKeyboard = true;
    } );

    $('#screenshot').mouseleave( function() {
        console.log("Mouse left screen")
        CaptureKeyboard = false;
    } );

    $('#instruction-area').mouseenter( function() {
       // console.log("Mouse entered instruction area")
        CaptureCPUControl = true;
    } );

    $('#instruction-area').mouseleave( function() {
        //console.log("Mouse left instruction area")
        CaptureCPUControl = false;
    } );

    $('#force-aux').change(function (){
        publishMemory();
    });

    $('#break-action').change(function (){
        updateBreakAction();
    });

    // document.addEventListener('keydown', function(event) {
    //     console.log("key pressed "+JSON.stringify(event));
    //     SendCommand("sendkey", [String(event.keyCode)]);
    // }, true);

    //updateScreenMS(1000);

    // $('#instruction-area').keydown(function (event) {
    //     var code = event.key;

    //     // S - step
    //     if (code == "s" && event.ctrlKey) {
    //         stepCPUOut();
    //     } else if (code == "s" && !event.ctrlKey) {
    //         stepCPU();
    //     } else if (code == "S") {
    //         stepCPUOver();
    //     } else if (code == "c") {
    //         runCPU();
    //     } else if (code == "p") {
    //         pauseCPU();
    //     } else if (code == "t") {
    //         traceOn();
    //     } else if (code == "T") {
    //         traceOff();
    //     }
    // });

    $('body').keydown(function (event) {

        if (CaptureCPUControl) {

            //console.log("captured")

            // respond to simple commands here
            var code = event.key;

            // S - step
            if (code == "s" && event.ctrlKey) {
                stepCPUOut();
            } else if (code == "s" && !event.ctrlKey) {
                stepCPU();
            } else if (code == "S") {
                stepCPUOver();
            } else if (code == "c") {
                runCPU();
            } else if (code == "p") {
                pauseCPU();
            } else if (code == "t") {
                traceOn();
            } else if (code == "T") {
                traceOff();
            }

            return;
        }

        if (!CaptureKeyboard) {
            return;
        }

        switch (event.key) {
        case "Backspace":
            SendCommand('sendkey', [String(127)]);
            break;
        case "CapsLock":
            break;
        case "ArrowDown":
            SendCommand('sendkey', [String(10)]);
            break;
        case "ArrowUp":
            SendCommand('sendkey', [String(11)]);
            break;
        case "ArrowLeft":
            SendCommand('sendkey', [String(8)]);
            break;
        case "ArrowRight":
            SendCommand('sendkey', [String(21)]);
            break;
        case "Escape":
            SendCommand('sendkey', [String(27)]);
            break;
        case "Alt":
            break;
        case "Meta":
            break;
        case "Shift":
            break;
        case "Ctrl":
            break;
        case "Enter":
            SendCommand('sendkey', [String(13)]);
            break;
        default:
            var code = event.key.charCodeAt(0);
            console.log("key "+event.key);
            if (event.ctrlKey && code >= 97 && code <= 120) {
                code -= 96
                console.log("ctrl code -> "+code);
                SendCommand( "sendkey", [ String(code) ] );
            } else {
                console.log(code);
                SendCommand( "sendkey", [ String(code) ] );
            }
        }
    });

    $('#heatmap-mode').change( function () {
        updateHeatmapMode();
    });

    $('#heatmap-bank').change( function () {
        updateHeatmapMode();
    });

    $('#heatmap-bank').keypress( function (event) {
        if (event.keyCode == 13) {
            updateHeatmapMode();
            event.preventDefault();
        };
    });

    // img.onload = function(ev) {
    //     console.log("Drawing image");
    //     var canvas = document.getElementById("screenshot");
    //     var context = canvas.getContext("2d");   
    //     context.drawImage(img, 0, 0);     
    // };
    $('#exec-address').keypress( function (event) {
        if (event.keyCode == 13) {
            var addr = $('#exec-address').val();
            SendCommand( "setpc", [ String(addr) ] );
            $('#go-dropdown').foundation('close');
            return false;
        } else if (event.keyCode == 27) {
            $('#go-dropdown').foundation('close');
        }
    });

    $('#memory-view-address').keypress( function (event) {
        if (event.keyCode == 13) {
            var addr = $('#memory-view-address').val();
            debugState.watch.address = addr;
            publishMemory();
            $('#mem-dropdown').foundation('close');
            return false;
        } else if (event.keyCode == 27) {
            $('#mem-dropdown').foundation('close');
        }
    });

    $('#mem-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        var addr = $('#memory-view-address').val();
        debugState.watch.address = addr;
        publishMemory();
        $('#mem-dropdown').foundation('close');
        return false;
    });

    $('#go-dropdown').submit(function(e){
        console.log("Doing submit...");
        e.preventDefault();
        var addr = $('#exec-address').val();
        SendCommand( "setpc", [ String(addr) ] );
        $('#go-dropdown').foundation('close');
        return false;
    });

    window.onfocus = function() {
        console.log("Focus");
        active = true;
    }

    window.onblur = function() {
        console.log("Blur");
        active = false;
    }


    SendCommand('get-settings', []);

    var targetvm = getUrlParameter("attach");
    if (targetvm != "") {
        var vm = targetvm - 1
        console.log("Request to attach to slot "+vm);
        attach(vm);
    }

}

function message(kind, title, message) {

    $('#footermessage').html(
        '<p class='+kind+'><b>' + title + '</b>: ' +
        message + '</p>'
    );

    // setTimeout( function() {
    //     $('#footermessage').html("");
    // }, 2000);

}

function getBreakpointArgs() {
    var pc = $('#reg-pc').val();
    var a = $('#reg-a').val();
    var x = $('#reg-x').val();
    var y = $('#reg-y').val();
    var sp = $('#reg-sp').val();
    var wa = $('#write-addr').val();
    var wv = $('#write-value').val();
    var read = $('#mem-action-read').is(":checked");
    var write = $('#mem-action-write').is(":checked");
    var bankMain = ($('#mem-bank-trap').val() === "M");
    var bankAux = ($('#mem-bank-trap').val() === "A");
    var fN = $('#flag-n').is(":checked");
    var fV = $('#flag-v').is(":checked");
    var fB = $('#flag-b').is(":checked");
    var fD = $('#flag-d').is(":checked");
    var fI = $('#flag-i').is(":checked");
    var fZ = $('#flag-z').is(":checked");
    var fC = $('#flag-c').is(":checked");
    var action = $('#break-action').val();
    var actionText = $('#break-text').val();
    var actionAddr = $('#break-jump-addr').val();
    var actionSpeed = $('#break-speed').val();
    var maxCount = $('#break-on-n').val();

    var params = ""

    var p = 0
    if (fN === true) {
        p |= flagN
    }
    if (fV === true) {
        p |= flagV
    }
    if (fB === true) {
        p |= flagB
    }
    if (fD === true) {
        p |= flagD
    }
    if (fI === true) {
        p |= flagI
    }
    if (fZ === true) {
        p |= flagZ
    }
    if (fC === true) {
        p |= flagC
    }

    if (pc !== "") {
        params += sprintf("PC=%s ",pc)
    }

    if (a !== "") {
        params += sprintf("A=%s ",a)
    }

    if (x !== "") {
        params += sprintf("X=%s ",x)
    }

    if (y !== "") {
        params += sprintf("Y=%s ",y)
    }

    if (sp !== "") {
        params += sprintf("SP=%s ",sp)
    }

    if (read === true) {
        if (wa !== "") {
            params += sprintf("RA=%s ",wa)
        }

        if (wv !== "") {
            params += sprintf("RV=%s ",wv)
        }
    } else if (write === true) {
        if (wa !== "") {
            params += sprintf("WA=%s ",wa)
        }

        if (wv !== "") {
            params += sprintf("WV=%s ",wv)
        }      
    }

    if (wa !== "") {
        if (bankAux === true) {
            params += "AUX=1 "
        }
        if (bankMain === true) {
            params += "MAIN=1"
        }
    }

    if (p !== 0) {
        params += sprintf("P=$%02x ",p)
    }

    if (maxCount !== "") {
        params += sprintf("MAXCNT=%s ", maxCount)
    } 

    if (action != null && action != "0") {
        // encode action 
        params += sprintf("ACTID=%s ", action);
        if (action == "1" || action == "8") {
            params += "ACTTXT="+actionText.replace( / /g, "_" ) + " ";
        } else if (action == "5") {
            params += sprintf("ACTJMP=%s ", actionAddr);
        } else if (action == "7") {
            params += sprintf("ACTSPD=%s ", actionSpeed);
        }
    }

    var args = params.trim().split(" ");
    console.log(JSON.stringify(args));

    return args
}

function clearCounter(idx) {
    SendCommand("setbpcounter", [String(idx-1), "0"]);
}

function refreshBreakpoints() {
    if (!$('#panel-breakpoints').hasClass('is-active')) {
        return;
    }
    SendCommand("listbp", [""]);
}

function updateCounters( r ) {
    console.log("counter update: "+JSON.stringify(r));
    var id = r['BreakpointID'];
    var value = r['Counter'];
    console.log("counter "+id+" is "+value);
    debugState.breakpoints[id]['Counter'] = value;
    decodeBreakpoints( { Breakpoints: debugState.breakpoints } );
}

function editBp(idx) {
    console.log("edit bp "+idx);
    if (idx == null || idx == "") {
        idx = "999";
    }
    var i = Number(idx) - 1;
    var bp = {};
    $('#bp-idx').val("");
    if (i >= 0 && i < debugState.breakpoints.length) {
        console.log("reading bp data idx "+i);
        bp = debugState.breakpoints[i];
        $('#bp-idx').val(i);
    }

    if (bp["MaxCount"] != null) {
        $('#break-on-n').val( sprintf( "%d", bp["MaxCount"] ) );
    } else {
        $('#break-on-n').val( "" );
    }
    if (bp["ValuePC"] != null) {
        $('#reg-pc').val( sprintf( "$%04x", bp["ValuePC"]) );
    } else {
        $('#reg-pc').val( "" );
    }
    if (bp["ValueA"] != null) {
        $('#reg-a').val( sprintf( "$%02x", bp["ValueA"] ) );
    } else {
        $('#reg-a').val( "" );
    }
    if (bp["ValueX"] != null) {
        $('#reg-x').val( sprintf( "$%02x", bp["ValueX"] ) );
    } else {
        $('#reg-x').val( "" );
    }
    if (bp["ValueY"] != null) {
        $('#reg-y').val( sprintf( "$%02x", bp["ValueY"] ) );
    } else {
        $('#reg-y').val( "" );
    }
    if (bp["ValueSP"] != null) {
        $('#reg-sp').val( sprintf( "$%04x", bp["ValueSP"] ) );
    } else {
        $('#reg-sp').val( "" );
    }
    if (bp["ValueP"] != null) {
        var pp = Number(bp['ValueP']);
        $('#flag-n').prop("checked", (pp & 0x80));
        $('#flag-v').prop("checked", (pp & 0x40));
        $('#flag-b').prop("checked", (pp & 0x10));
        $('#flag-d').prop("checked", (pp & 0x08));
        $('#flag-i').prop("checked", (pp & 0x04));
        $('#flag-z').prop("checked", (pp & 0x02));
        $('#flag-c').prop("checked", (pp & 0x01));
    } 
    $('#write-addr').val( "" );
    $('#write-value').val( "" );
    if (bp["WriteAddress"] != null) {
        $('#write-addr').val( sprintf( "$%04x", bp["WriteAddress"] ) );
        $('#mem-action-write').prop("checked", true);
    } 
    if (bp["WriteValue"] != null) {
        $('#write-value').val( sprintf( "$%02x", bp["WriteValue"] ) );
    } 
    if (bp["ReadAddress"] != null) {
        $('#write-addr').val( sprintf( "$%04x", bp["ReadAddress"] ) );
        $('#mem-action-read').prop("checked", true);
    } 
    if (bp["ReadValue"] != null) {
        $('#write-value').val( sprintf( "$%02x", bp["ReadValue"] ) );
    } 
    console.log("main "+bp['Main']);
    console.log("aux "+bp['Aux']);
    if (bp['Main'] === 1) {
        // $('#mem-bank-main').prop("checked", true);
        $("#mem-bank-trap option[value='M']").prop( 'selected', true );
    } else if (bp['Aux'] === 1) {
        // $('#mem-bank-aux').prop("checked", true);
        $("#mem-bank-trap option[value='A']").prop( 'selected', true );
    } else {
        $("#mem-bank-trap option[value='*']").prop( 'selected', true );
    }
    console.log(bp['Action'])
    $('#break-text').prop("disabled", true);
    $('#break-speed').prop("disabled", true);
    $('#break-jump-addr').prop("disabled", true);
    if (bp['Action'] != null) {
        // There is an action here -- decode it!
        $("#break-action option[value='"+bp['Action']['Type']+"']").prop( 'selected', true );
        if (bp['Action']['Type'] == "1" || bp['Action']['Type'] == "8") {
            $('#break-text').val( bp['Action']['Arg0'] );
            $('#break-text').prop("disabled", false);
        } else if (bp['Action']['Type'] == "5") {
            $('#break-jump-addr').val( sprintf( "$%04x", bp['Action']['Arg0'] ) );
            $('#break-jump-addr').prop("disabled", false);
        } else if (bp['Action']['Type'] == "7") {
            $("#break-speed option[value='"+bp['Action']['Arg0']+"']").prop( 'selected', true );
            $('#break-speed').prop("disabled", false);
        }
    } else {
        $('#break-action').val( "0" );
        $('#break-text').val( "" );
        $('#break-jump-addr').val( "" );
        $('#break-speed option[value="1"]').prop( 'selected', true );
    }

    // show popup
    $('#add-bp-dropdown').foundation('open');
    $('#reg-pc').focus();
}

function addBreakpoint() {

    var args = getBreakpointArgs();
    var idx = $('#bp-idx').val();

    if (idx != null && idx != "") {
        args.unshift( idx );
        SendCommand( "updbp", args );
    } else {    
        SendCommand( "setbp", args );
    }

    return false
}

function updateLRState( lr ) {

    $('#live-rewind').prop('checked', (lr['Enabled'] === true));

    if (lr['Enabled'] === false && lr['CanResume'] === false && lr['CanBack'] === false && lr['CanForward'] === false) {
        $('#lr-resume').prop('disabled', true);
        $('#lr-resume-paused').prop('disabled', true);     
        $('#lr-back').prop('disabled', true);  
        $('#lr-start').prop('disabled', true); 
        $('#lr-forward').prop('disabled', true);
        $('#lr-play').prop('disabled', true);
        $('#lr-pause').prop('disabled', true);
        $('#lr-skip1b').prop('disabled', true);
        $('#lr-skip10b').prop('disabled', true);
        $('#lr-skip1f').prop('disabled', true);
        $('#lr-skip10f').prop('disabled', true);
        $('#lr-skip100b').prop('disabled', true);
        $('#lr-skip1000b').prop('disabled', true);
        $('#lr-skip100f').prop('disabled', true);
        $('#lr-skip1000f').prop('disabled', true);
        message("success", "Recording", "Recording is disabled (update LR State)");
        updateCPUStatus( debugState.status );
        return
    }

    if (Number(lr['TotalBlocks']) > 0) {
        $('#play-position').html("<button class='button tiny clear'>"+lr['Block']+"/"+lr['TotalBlocks']+"</button>");
    }

    if (lr['CanResume'] === true) {
        $('#lr-resume').prop('disabled', false);
        $('#lr-resume-paused').prop('disabled', false);
    } else {
        $('#lr-resume').prop('disabled', true);
        $('#lr-resume-paused').prop('disabled', true);
    }

    if (lr['CanBack'] === true) {
        $('#lr-back').prop('disabled', false);
        $('#lr-start').prop('disabled', false);
    } else {
        $('#lr-back').prop('disabled', true);
        $('#lr-start').prop('disabled', true);
    }

    if (lr['CanForward'] === true) {
        $('#lr-forward').prop('disabled', false);
        $('#lr-play').prop('disabled', false);
        $('#lr-pause').prop('disabled', false);
        $('#lr-start').prop('disabled', false); 
        //
        $('#cpu-step').prop('disabled', true);
        $('#cpu-stepo').prop('disabled', true);
        $('#cpu-stepx').prop('disabled', true);
        $('#cpu-pause').prop('disabled', true);
        $('#cpu-continue').prop('disabled', true); 
        $('#cpu-go').prop('disabled', true);
        $('#cpu-trace').prop('disabled', true);
        $('#cpu-trace-off').prop('disabled', true);
        $("#cpu-speed-025x").prop('disabled', true);
        $("#cpu-speed-05x").prop('disabled', true);
        $("#cpu-speed-1x").prop('disabled', true);
        $("#cpu-speed-2x").prop('disabled', true);
        $("#cpu-speed-4x").prop('disabled', true);

        // update play pos
        // if (Number(lr['TotalBlocks']) > 0) {
        //     $('#play-position').html("<button class='button tiny clear'>"+lr['Block']+"/"+lr['TotalBlocks']+"</button>");
        // }
    } else {
        $('#lr-forward').prop('disabled', true);
       // $('#lr-start').prop('disabled', true); 
        $('#lr-play').prop('disabled', true);
        $('#lr-pause').prop('disabled', false);
        //
        $('#play-position').html("");
    }

    var msg = '';

    if (lr['CanForward'] === true) {
        if (lr['TimeFactor'] === 0) {
            msg = "Paused"
            message("success", "Recording", "Paused");
            $('#lr-skip1b').prop('disabled', false);
            $('#lr-skip10b').prop('disabled', false);
            $('#lr-skip1f').prop('disabled', false);
            $('#lr-skip10f').prop('disabled', false);
            $('#lr-skip100b').prop('disabled', false);
            $('#lr-skip1000b').prop('disabled', false);
            $('#lr-skip100f').prop('disabled', false);
            $('#lr-skip1000f').prop('disabled', false);
            return
        } else {
            $('#lr-skip1b').prop('disabled', true);
            $('#lr-skip10b').prop('disabled', true);
            $('#lr-skip1f').prop('disabled', true);
            $('#lr-skip10f').prop('disabled', true);
            $('#lr-skip100b').prop('disabled', true);
            $('#lr-skip1000b').prop('disabled', true);
            $('#lr-skip100f').prop('disabled', true);
            $('#lr-skip1000f').prop('disabled', true);
        }
        if (lr['Backwards'] === true) {
            msg += "Rew "
        } else {
            msg += "Fwd "
        }
        msg += sprintf( "%02fx", lr['TimeFactor'] );
        message("success", "Recording", msg);
    } else {
        message("success", "Recording", "Recording is enabled");
        //console.log("Updating CPU status to "+debugState.status)
        updateCPUStatus( debugState.status );
    }

}

function toggleRewind() {
    var rewind = $('#live-rewind').is(":checked");
    console.log("Updating rewind state to "+rewind);
    SendCommand( "live-rewind", [ "set", String(rewind) ] );
}

function lrResume() {
    SendCommand( "live-rewind", [ "resume" ] );
}

function lrPause() {
    SendCommand( "live-rewind", [ "pause" ] );
}

function lrResumePaused() {
    SendCommand( "live-rewind", [ "resume-cpu-paused" ] );
}

function lrBack() {
    SendCommand( "live-rewind", [ "back" ] );
}

function lrStart() {
    SendCommand( "live-rewind", [ "reset" ] );
}

function lrForward() {
    SendCommand( "live-rewind", [ "forwards" ] );
}

function lrForward1x() {
    SendCommand( "live-rewind", [ "forwards1x" ] );
}

function lrSkip(amt) {
    SendCommand( "live-rewind", [ "jump", String(amt) ] );
}

function setWarp(amt) {
    console.log("Setting warp to "+amt);
    SendCommand( "setwarp", [ String(amt) ] );
}

function lrSkip1s() {
    SendCommand( "live-rewind", [ "jump", "10" ] );
}

function lrSkip100ms() {
    SendCommand( "live-rewind", [ "jump", "1" ] );
}

function icon(name, fill) {
    if (fill == null) {
        fill = "#000000";
    }
    return sprintf("<span style='content: url(/svg/fi-%s.svg); width: 20px; height 20px; margin-left: 10px; margin-right: 10px;'></span>", name, fill);
}

function assembleCode() {
    // Get the assembly code from the textarea
    var asmCode = $('#assembler-text').val();

    // Base64 encode the code
    var base64Code = btoa(asmCode);

    // Prepare the request payload
    var payload = {
        files: [
            {
                name: "main.s",
                data: base64Code,
                binary: false
            }
        ]
    };

    // Make the POST request
    $.ajax({
        url: 'https://turtlespaces.org:6502/api/v1/asm/multifile',
        type: 'POST',
        contentType: 'application/json',
        data: JSON.stringify(payload),
        success: function(response) {
            // Parse the JSON response string
            var parsedResponse;
            try {
                parsedResponse = JSON.parse(response);
            } catch (e) {
                alert("Failed to parse response: " + e.message + "\n\nRaw response:\n" + response);
                return;
            }

            // Check if there are errors
            if (parsedResponse.err && parsedResponse.err.length > 0) {
                // Handle errors
                var errorMsg = "Assembly error(s) occurred:\n\n";
                parsedResponse.err.forEach(function(error) {
                    console.log("Error object:", error);  // Debug log
                    // Workaround for server bug where filename is set to line number
                    var filename = error.filename;
                    var lineNumber = error.line;
                    if (/^\d+$/.test(filename)) {
                        // If filename is just digits, it's likely the line number
                        lineNumber = parseInt(filename);
                        filename = "main.s";
                    }
                    errorMsg += "Line " + lineNumber + " in " + filename + ": " + error.message + "\n";
                });
                alert(errorMsg);
            } else if (parsedResponse.data && parsedResponse.addr !== undefined) {
                // Success - assembled code
                // Decode base64 data
                var decodedData = atob(parsedResponse.data);
                var bytes = [];
                for (var i = 0; i < decodedData.length; i++) {
                    bytes.push(decodedData.charCodeAt(i));
                }

                var length = bytes.length;
                var address = parsedResponse.addr;
                var filename = parsedResponse.name || "unnamed";
                var successMsg = filename + ": " + length + " bytes assembled to memory address $" + address.toString(16).toUpperCase();

                // If less than 32 bytes, show the hex dump
                if (length < 32) {
                    var hexBytes = bytes.map(function(b) {
                        return ("0" + b.toString(16)).slice(-2).toUpperCase();
                    }).join(" ");
                    successMsg += "\n\nHex bytes: " + hexBytes;
                }
                alert(successMsg);

                // Update memory with the assembled bytes
                var binaryData = new Uint8Array(bytes);
                const options = {
                    method: 'post',
                    headers: {
                        'Content-type': 'application/octet-stream; charset=UTF-8',
                        'X-LoadAddress': String(address),
                    },
                    body: binaryData,
                };

                fetch("/api/debug/upload", options).then(response => {
                    console.log("Memory updated at address $" + address.toString(16).toUpperCase());
                }).catch(err => {
                    console.error('Failed to update memory', err);
                    alert('Warning: Assembly succeeded but failed to update memory: ' + err);
                });
            } else {
                alert("Unexpected response format from assembly service:\n\n" + JSON.stringify(parsedResponse, null, 2));
            }
        },
        error: function(xhr, status, error) {
            alert("Failed to assemble code: " + error);
        }
    });
}

$( document ).ready(function() {
    // Handler for .ready() called.
    initialize();
});
