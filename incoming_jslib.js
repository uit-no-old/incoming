_incoming_lib = function() {
    var incoming_lib = {};

    var server_hostname = null;
    incoming_lib.set_server_hostname = function set_server_hostname(hostname) {
        server_hostname = hostname;
    };

    var msgUploadReq = function msgUploadReq(upload_id, length_bytes) {
        var msg = {
            MsgType: "MsgUploadReq",
            MsgData : {
                Id: upload_id,
                LengthBytes: length_bytes
            }
        };
        return JSON.stringify(msg);
    };

    var msgAck = function msgAck(ack) {
        var msg = {
            MsgType: "MsgAck",
            MsgData: {
                Ack: ack
            }
        };
        return JSON.stringify(msg);
    };

    var msgCancelAck = function msgCancelAck(ack) {
        var msg = {
            MsgType: "MsgCancelAck",
            MsgData: {
                Ack: ack
            }
        };
        return JSON.stringify(msg);
    };

    var msgPause = function msgPause() {
        var msg = {
            MsgType: "MsgPause",
            MsgData: {
                Pause: true
            }
        };
        return JSON.stringify(msg);
    };

    var msgCancel = function msgCancel(reason) {
        var msg = {
            MsgType: "MsgCancel",
            MsgData: {
                Reason: reason
            }
        };
        return JSON.stringify(msg);
    };

    var msgError = function msgError(reason) {
        var msg = {
            MsgType: "MsgError",
            MsgData: {
                ErrorCode: 0,
                Msg: reason
            }
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
        var conn_retry = null;
        ul.chunks_tx_now = 0; // "now" because upload could have been resumed
        ul.chunks_acked_now = 0; // "now" because upload could have been resumed
        ul.bytes_tx = 0; // bytes sent over websocket
        ul.bytes_acked = 0; // bytes acknowledged by incoming backend
        ul.bytes_total = file.size;
        ul.finished = false;
        ul.cancelled = false;
        ul.cancelling = false;
        ul.can_cancel = true;
        ul.can_pause = false;
        ul.connected = false;
        ul.paused = false;
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

        // the following callbacks should be set by the caller directly. All are
        // functions taking one parameter: the uploader object.
        ul.onprogress = function(o){};
        ul.onfinished = function(o){};
        ul.oncancelled = function(o){};
        ul.onerror = function(o){};

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
                    ul.bytes_tx < ul.bytes_total &&
                    !ul.cancelling) {
                var end = ul.bytes_tx + (upload_conf.ChunkSizeKB*1024);
                if (end > ul.bytes_total) {
                    end = ul.bytes_total;
                }
                var blob = file.slice(ul.bytes_tx, end);
                file_reader.readAsArrayBuffer(blob);
            }
        };

        // file_reader - when a chunk is loaded, send it over the websocket,
        // then try to load&send another chunk. If something went wrong
        // during loading, cancel the upload.
        file_reader.onload = function onload(evt) {
            if (evt.target.readyState == FileReader.DONE &&
                    evt.target.result != null &&
                    !ul.cancelling && !ul.cancelled && !ul.paused) {
                // send chunk if websocket is open
                if (ws.readyState == WebSocket.OPEN) {
                    buf = evt.target.result;
                    ws.send(buf);

                    // update state
                    ul.bytes_tx += buf.byteLength;
                    ul.bytes_ahead += buf.byteLength;
                    ul.chunks_tx_now += 1;
                    ul.chunks_ahead += 1;

                    // if this was the last chunk, we can no longer
                    // cancel or pause
                    if (ul.bytes_tx == ul.bytes_total) {
                        ul.can_cancel = false;
                        ul.can_pause = false;
                    }

                    // call progress cb
                    ul.onprogress(ul);

                    // try to load&send another chunk
                    try_load_and_send_file_chunk();
                }
            } else {
                if (!ul.cancelling && !ul.cancelled && !ul.paused) {
                    ul.cancel("unexpected file_reader.onloadend error");
                }
            }
        };
        file_reader.onerror = function onerror(evt) {
            if (!ul.cancelled) {
                ul.cancel("error on file load: " + evt.name + " " + evt.message);
            }
        };

        // when we receive an ack from the incoming backend that a chunk has
        // been received, we can try to load&send another chunk. When all
        // chunks have been sent and acked, we are finished reading chunk acks.
        // (Instead, we will wait for the final "upload is done" message)
        var receive_chunk_acks = function receive_chunk_acks(msg) {
            // after handshake is done in start(), this is ws.onmessage.
            var obj = JSON.parse(msg.data);
            if (obj.MsgType == "MsgChunkAck") {
                // update state
                ul.bytes_acked += obj.MsgData.ChunkSize;
                ul.bytes_ahead -= obj.MsgData.ChunkSize;
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

            } else if (obj.MsgType == "MsgError") {
                ul.error_code = obj.MsgData.ErrorCode;
                ul.error_msg = obj.MsgData.Msg;
                ul.onerror(ul);
                ul.cancel("Error from server: " + ul.error_msg);
            } else if (obj.MsgType == "MsgCancel") {
                ul.cancel(obj.MsgData.Reason);
            } else {
                alert("Bug! Didn't understand what came out of the socket");
            }
        };

        // when all is uploaded, we wait for the incoming!! backend to tell us
        // that the upload is done; that even the web app backend is done
        // fetching the file from the incoming!! backend.
        var receive_final_message = function receive_final_message(msg) {
            var obj = JSON.parse(msg.data);
            if (obj.MsgType == "MsgAllDone") {
                ul.finished = true;
                ws.close();
                ul.state_msg = "all done";
                ul.onfinished(ul);
            } else if (obj.MsgType == "MsgError") {
                ul.error_code = obj.MsgData.ErrorCode;
                ul.error_msg = obj.MsgData.Msg;
                ws.close();
                ul.state_msg = "Upload failed on server side: " + ul.error_msg;
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
            ul.connected = false;
            ul.paused = false;
            ul.can_cancel = true;
            ul.can_pause = false;
            ul.state_msg = "connecting to upload server";

            // open websocket TODO make sure it's the same security as the whole page (encrypted or non-encrypted)
            ws = new WebSocket('ws://' + server_hostname + '/frontend/upload_ws');
            //ws.binaryType = "arraybuffer"; // relevant only for recv? http://dev.w3.org/html5/websockets/#dom-websocket-binarytype

            ws.onopen = function onopen() {
                // we handle the handshake here, up until we start
                // sending file chunks.
                ul.connected = true;
                ul.state_msg = "upload protocol handshake"

                // send upload request
                ws.send(msgUploadReq(upload_id, file.size));

                // receive error or upload config
                ws.onmessage = function prot01_recvConfig(msg) {
                    var obj = JSON.parse(msg.data);
                    if (obj.MsgType == "MsgError") {
                        ul.error_code = obj.MsgData.ErrorCode;
                        ul.error_msg = obj.MsgData.Msg;
                        ul.onerror(ul);
                        ul.cancel("can't handle '" + ul.error_msg + "'");
                    } else if (obj.MsgType == "MsgUploadConf") {
                        // got upload config. set us up for upload!
                        upload_conf = obj.MsgData;
                        ul.bytes_acked = upload_conf.FilePos;
                        ul.bytes_tx = ul.bytes_acked;
                        ws.send(msgAck(true));
                        // we might already be finished uploading data...
                        if (ul.bytes_acked == ul.bytes_total) {
                            ul.can_cancel = false;
                            ul.can_pause = false;
                            ul.state_msg = "processing file on server";
                            ws.onmessage = receive_final_message;
                        } else {
                            ul.state_msg = "transfer file chunks to upload server"
                            ul.can_pause = true;
                            ws.onmessage = receive_chunk_acks;
                            // start uploading chunks
                            try_load_and_send_file_chunk();
                        }
                        ul.onprogress(ul);
                    } else {
                        alert("Bug! Didn't understand what came out of the socket");
                    }
                };
            };

            ws.onclose = function onclose(msg) {
                ul.connected = false;
                if (ul.paused || ul.finished || ul.cancelled || ul.error_code != null) {
                    ul.onprogress(ul);
                    ws = null;
                    return;
                }

                // try to start again every 20 seconds
                ul.state_msg = "lost connection, trying to start again every 20 seconds";
                ul.connected = false;
                conn_retry = setTimeout(ul.start, 20000);
                ws = null;
                ul.can_pause = true;
                ul.onprogress(ul);
                // The event object passed to "onclose" has three fields named
                // "code", "reason", and "wasClean".  The "code" is a numeric
                // status provided by the server, and can hold the same values
                // as the "code" argument of close().  The "reason" field is a
                // string describing the circumstances of the "close" event.
            }

            ws.onerror = function onerror(evt_error) {
                ws.close();
                //ul.error_code = 0;
                //ul.error_msg = "websocket error: " + evt_error.message;
                //ul.onerror(ul);
                //ul.cancel("websocket error: " + evt_error.message);
            }
            return ul;
        };

        ul.cancel = function cancel(reason) {
            if (!ul.cancelled && ul.can_cancel) {
                file_reader.abort()
                if (ws != null && ws.readyState == WebSocket.OPEN) {
                    ul.cancelling = true;
                    ws.send(msgCancel(reason));

                    // we need the backend to ack this, so we set ws.onmessage to
                    // receive a MsgCancelAck. However, there might be other
                    // messages arriving before that (most likely chunk acks),
                    // so we need to call whichever "usual" message handler we have
                    // in ws.onmessage in case we don't get a MsgCancelAck.
                    var usual_onmessage_handler = ws.onmessage;
                    ws.onmessage = function recv_cancel_ack(msg) {
                        var obj = JSON.parse(msg.data);
                        if (obj.MsgType == "MsgCancelAck") {
                            ul.cancelled = true;
                            ul.cancelling = false;
                            ul.can_cancel = false;
                            ul.state_msg = "cancelled: " + reason;
                            ul.cancel_msg = reason;
                            ws.close();
                            ul.oncancelled(ul);
                        } else {
                            // we have received a different message - most
                            // likely chunk acks. Call message handler that we
                            // would call normally.
                            usual_onmessage_handler(msg);
                            // message handler might have changed ws.onmessage
                            if (ws.onmessage != recv_cancel_ack) {
                                usual_onmessage_handler = ws.onmessage;
                                ws.onmessage = recv_cancel_ack;
                            }
                        }
                    };
                } else { 
                    // we couldn't send a cancel message to the backend
                    ul.cancelled = true;
                    ul.can_cancel = false;
                    ul.state_msg = "cancelled: " + reason;
                    ul.cancel_msg = reason;
                    ul.error_code = 0;
                    ul.error_msg = "upload is cancelled, but the backend doesn't know it yet";
                    ul.onerror(ul);
                    ul.oncancelled(ul)
                }

                // the following regardless of whether we could send a cancel
                // message to the backend or not
                ul.can_cancel = false;
                ul.can_pause = false;
                ul.state_msg = "cancelling: " + reason;
                if (conn_retry != null) {
                    clearTimeout(conn_retry);
                    conn_retry = null;
                }
                ul.onprogress(ul);
            }
            return ul;
        };

        // pause pauses, unpauses, or toggles pause, depending on the
        // parameter. "pause" pauses, "unpause" unpauses, "toggle" toggles.
        ul.pause = function pause(what) {
            if (ul.finished || ul.cancelled) {
                return;
            }

            var new_state = false;
            if (what == "pause") {new_state = true;}
            else if (what == "unpause") {new_state = false;}
            else if (what == "toggle") {new_state = !ul.paused;}
            else {alert("Bug! Uploader.pause() called with unknown parameter!");}

            if (new_state == ul.paused) {
                return;
            }

            if (conn_retry != null) {
                clearTimeout(conn_retry);
                conn_retry = null;
            }

            if (new_state) {
                // pause upload
                file_reader.abort();
                if (ws != null && ws.readyState == WebSocket.OPEN) {
                    ws.send(msgPause());
                    ws.close();
                }
                ul.paused = true;
                ul.can_cancel = false;
                ul.state_msg = "paused";
                ul.onprogress(ul);
            } else {
                // unpause upload: call ul.start(), which should connect and resume.
                ul.paused = false;
                ul.state_msg = "unpaused";
                ul.onprogress(ul);
                ul.start();
            }
        };

        return ul;
    };

    return incoming_lib;
}

incoming = _incoming_lib();
