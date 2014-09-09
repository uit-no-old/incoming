_incoming_lib = function() {
    var incoming_lib = {};

    var server_hostname = null;
    incoming_lib.set_server_hostname = function set_server_hostname(hostname) {
        server_hostname = hostname;
    };

    var msgUploadReq = function msgUploadReq(upload_id, length_bytes) {
        var msg = {
            Id: upload_id,
            LengthBytes: length_bytes
        };
        return JSON.stringify(msg);
    };

    var msgAck = function msgAck(ack) {
        var msg = {
            Ack: ack
        };
        return JSON.stringify(msg);
    };

    incoming_lib.Uploader = function Uploader(upload_id, file) {
        var ul = {};

        var ws = null; // WebSocket - opened in start()
        var file_reader = new FileReader();
        var upload_conf = null; // from incoming - has ChunkSizeKB,
                                // FilePos (for resume), SendAhead.
                                // Set in start()
        ul.chunks_tx_now = 0; // "now" because upload could have been resumed
        ul.chunks_acked_now = 0; // "now" because upload could have been resumed
        ul.bytes_tx = 0; // bytes sent over websocket
        ul.bytes_acked = 0; // bytes acknowledged by incoming backend
        ul.bytes_total = file.size;
        ul.finished = false;
        ul.cancelled = false;
        ul.chunks_ahead = 0; // how many chunks have been sent without
                             // having gotten acknowledgement from incoming
                             // (upload_conf.SendAhead is upper limit for this)
        ul.bytes_ahead = 0;  // how many bytes have been sent without
                             // having gotten acknowledgement from incoming
        ul.error_code = null; // set only in case of error
        ul.error_msg = null;
        ul.cancel_msg = null;
        ul.state_msg = "not yet started"; // purely informal, human readable state
                                          // information

        // the following callbacks should be set by the caller directly
        ul.onprogress = null;
        ul.onfinished = null;
        ul.oncancelled = null;
        ul.onerror = null;

        // try_load_and_send_file_chunk is the function we call when we want to
        // send chunks. It makes file_reader load a chunk from the file. When
        // that chunk is loaded, file_reader will send it over the websocket
        // and call this function again (see onloadend below). In consequence,
        // this function will be called even if we can't send anything at the
        // moment. We call it
        //  - whenever we have just sent a chunk in onloadend (because we might
        //  be able to send another one)
        //  -  when we receive an ack from the incoming backend that a chunk
        //  has been transferred (in receive_chunk_acks), and
        //  -  once in the beginning, in start().
        var try_load_and_send_file_chunk = function try_load_and_send_file_chunk() {
            // are we clear to transfer more right now? if yes, make
            // file_reader load (and in onloadend send) the next chunk
            if (ul.chunks_ahead < upload_conf.SendAhead && 
                    file_reader.readyState != FileReader.LOADING &&
                    ul.bytes_tx < ul.bytes_total) {
                var end = ul.bytes_tx + (upload_conf.ChunkSizeKB*1024);
                if (end > ul.bytes_total) {
                    end = ul.bytes_total;
                }
                var blob = file.slice(ul.bytes_tx, end);
                file_reader.readAsArrayBuffer(blob);
            }
        };

        // set up file_reader - when a chunk is loaded, send it over the websocket,
        // then try to load&send another chunk
        file_reader.onloadend = function onloadend(evt) {
            if (evt.target.readyState == FileReader.DONE &&
                    evt.target.result != null) {
                // send chunk
                buf = evt.target.result;
                ws.send(buf);

                // update state
                ul.bytes_tx += buf.byteLength;
                ul.bytes_ahead += buf.byteLength;
                ul.chunks_tx_now += 1;
                ul.chunks_ahead += 1;

                // call progress cb
                ul.onprogress(ul);

                // try to load&send another chunk
                try_load_and_send_file_chunk();
            } else {
                // TODO: error handling. file could be gone or something
                // onloadend is also called after abort()?
                alert("file_reader.onloadend error")
            }
        };

        // when we receive an ack from the incoming backend that a chunk has
        // been received, we can try to load&send another chunk. When all
        // chunks have been sent and acked, we are finished reading chunk acks.
        // (Instead, we will wait for the final "upload is done" message)
        var receive_chunk_acks = function receive_chunk_acks(msg) {
            // after handshake is done in start(), this is ws.onmessage.
            var obj = JSON.parse(msg.data);
            if (obj.hasOwnProperty("ChunkSize")) {
                // update state
                ul.bytes_acked += obj.ChunkSize;
                ul.bytes_ahead -= obj.ChunkSize;
                ul.chunks_acked_now += 1;
                ul.chunks_ahead -= 1;

                // if all file chunks have been acked, set ws.onmessage to wait
                // for final "upload complete" message. if not, send more
                // chunks
                if (ul.bytes_acked == ul.bytes_total) {
                    ul.state_msg = "processing file on server";
                    ws.onmessage = receive_final_message;
                } else {
                    try_load_and_send_file_chunk();
                }

                // call progress cb
                ul.onprogress(ul);

            } else if (obj.hasOwnProperty("ErrorCode")) {
                // error handling TODO probably more to do here? some sort of recovery? cancel ourselves?
                file_reader.abort();
                ul.error_code = obj.ErrorCode;
                ul.error_msg = obj.Msg;
                ul.onerror(ul);
            } else {
                alert("Bug! Didn't understand what came out of the socket");
            }
        };

        // when all is uploaded, we wait for the incoming!! backend to tell us
        // that the upload is done; that even the web app backend is done
        // fetching the file from the incoming!! backend.
        var receive_final_message = function receive_final_message(msg) {
            var obj = JSON.parse(msg.data);
            if (obj.hasOwnProperty("Success")) {
                ul.finished = true;
                ws.close();
                ul.state_msg = "all done";
                ul.onfinished(ul);
            } else if (obj.hasOwnProperty("ErrorCode")) {
                ul.error_code = obj.ErrorCode;
                ul.error_msg = obj.Msg;
                ul.onerror(ul);
            }
        };

        ul.start = function start() {
            // this is called on start and restarts too (when connection is lost).
            // So we (re)initialize some state first
            ul.chunks_tx_now = 0;
            ul.chunks_acked_now = 0;
            ul.bytes_tx = ul.bytes_acked;
            ul.chunks_ahead = 0;
            ul.bytes_ahead = 0;
            ul.state_msg = "connecting to upload server";

            // open websocket TODO make sure it's the same security as the whole page (encrypted or non-encrypted)
            ws = new WebSocket('ws://' + server_hostname + '/frontend/upload_ws');
            //ws.binaryType = "arraybuffer"; // relevant only for recv? http://dev.w3.org/html5/websockets/#dom-websocket-binarytype

            ws.onopen = function onopen() {
                // we handle the handshake here, up until we start
                // sending file chunks.
                ul.state_msg = "upload protocol handshake"

                // send upload request
                ws.send(msgUploadReq(upload_id, file.size));

                // receive error or upload config
                ws.onmessage = function prot01_recvConfig(msg) {
                    var obj = JSON.parse(msg.data);
                    if (obj.hasOwnProperty("ErrorCode")) {
                        ul.error_code = obj.ErrorCode;
                        ul.error_msg = obj.Msg;
                        ul.onerror(ul);
                        // TODO more error handling? cancel ourselves?
                    } else if (obj.hasOwnProperty("ChunkSizeKB")) {
                        // got upload config. set us up for upload!
                        upload_conf = obj;
                        ul.bytes_acked = upload_conf.FilePos;
                        ul.bytes_tx = ul.bytes_acked;
                        ws.onmessage = receive_chunk_acks;
                        ws.send(msgAck(true));
                        ul.state_msg = "transfer file chunks to upload server"
                        // start uploading chunks
                        try_load_and_send_file_chunk();
                    } else {
                        alert("Bug! Didn't understand what came out of the socket");
                    }
                };
            };

            ws.onclose = function onclose(msg) {
                // TODO implement! here or outside of start() (then just set here)
                // if not cancelled or error, try start() a few times until we give up
                // and error. Need to reset some internal state first?
                file_reader.abort();
                ul.state_msg = "connection to upload server closed"
                // The event object passed to "onclose" has three fields named "code", "reason", and "wasClean".  The "code" is a numeric status provided by the server, and can hold the same values as the "code" argument of close().  The "reason" field is a string describing the circumstances of the "close" event.
            }

            ws.onerror = function onerror(evt_error) {
                // TODO more error handling? try again or something?
                file_reader.abort();
                ws.close();
                alert(evt_error.message);
            }
            return ul;
        };
        return ul;
    };

    return incoming_lib;
}

incoming = _incoming_lib();
