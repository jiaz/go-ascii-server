var Debug = {
    InDebugMode: false,
    Log: function(str) {
        if (this.InDebugMode) {
            console.log(str);
        }
    }
}

/* Circular buffer works like an infinite queue, while the underlying queue size is limited.
 *
 * Operation supported:
 *
 *   - Enqueue: Push a new element to the end. When pushing to a full queue,
 *              the push will fail.
 *
 *   - Dequeue: Get a new element from the head and remove it from the buffer.
 *              When there is nothing in the queue, undefined will be returned.
 *
 * The buffer will also trigger `onChanged` callback when the content length is changed.
 */
function CircularBuffer(bufferSize) {
    this._bufferSize = bufferSize;
    this._buffer = [];
    this._buffer.length = bufferSize + 1; // one more space for sentinal
    this._sentinalPos = 0;
    this._tailPos = 0;

    this.onChanged = undefined;
}

CircularBuffer.prototype._nextPos = function(pos) {
    return (pos + 1) % (this._bufferSize + 1);
};

CircularBuffer.prototype.Enqueue = function(data) {
    var next = this._nextPos(this._tailPos);
    if (next == this._sentinalPos) {
        return false;
    }

    this._buffer[next] = data;
    this._tailPos = next;

    if (this.onChanged !== undefined) {
        this.onChanged(this);
    }

    return true;
};

CircularBuffer.prototype.Dequeue = function() {
    if (this._tailPos == this._sentinalPos) return undefined;
    else {
        var next = this._nextPos(this._sentinalPos, this._bufferSize);
        var res = this._buffer[next];
        this._sentinalPos = next;

        if (this.onChanged !== undefined) {
            this.onChanged(this);
        }

        return res;
    }
};

CircularBuffer.prototype.Length = function() {
    if (this._tailPos >= this._sentinalPos) {
        return this._tailPos - this._sentinalPos;
    } else {
        return (this._bufferSize - this._sentinalPos) + (this._tailPos + 1);
    }
};

function WSManager(host) {
    this._host = host;
    this._responseHandlerRegistry = {};
    this._errHandler;
    this._ws;
}

WSManager.prototype._hookupEvents = function(openCallback) {
    var self = this;
    this._ws.onmessage = function(ev) {
        var responseData = JSON.parse(ev.data)
        if (Debug.InDebugMode) {
        	Debug.Log("Received response, code:" + responseData.ErrorCode + ", type: " + responseData.Type);
        }
        if (responseData.ErrorCode !== 200) {
            self._errHandler(responseData);
        }

        var handlerFunc = self._responseHandlerRegistry[responseData.Type]
        if (handlerFunc === undefined) {
            Debug.Log("No handler defined for response type: " + responseData.Type);
        } else {
            handlerFunc(responseData.Data);
        }
    };

    this._ws.onopen = function(ev) {
    	Debug.Log("Connection opened.");
        openCallback(ev);
    }
}

WSManager.prototype.RegisterHandler = function(eventType, handler) {
    this._responseHandlerRegistry[eventType] = handler;
}

WSManager.prototype.SetErrorHandler = function(handler) {
    this._errHandler = handler;
}

WSManager.prototype.Connect = function(callback) {
    Debug.Log("Start websocket connection...");
    this._ws = new WebSocket("ws://" + this._host + "/play");
    this._hookupEvents(callback);
}

WSManager.prototype.SendCommand = function(cmdType, args) {
    var cmd = {
        type: cmdType
    };
    if (args !== undefined) {
        cmd.args = args;
    }
    this._ws.send(JSON.stringify(cmd));
}

function VideoPlayer(elm, host) {
    this.BUFFER_SIZE = 100; // buffer 30 seconds, 30 fps
    this.CACHE_LIMIT = 30; // cache 2 seconds before start playing
    this.FETCH_LIMIT = 30; // start fetch next buffer when remaining is less than 5 seconds

    this._wrapperElm = elm;
    this._frameElm = elm.find("#frame");
    this._notificationElm = elm.find("#notification");
    this._playBtn = elm.find("#play-btn");

    this._playingState = 0; // 0: stop, 1: allow to play
    this._isPlaying = 0; // 0: stop, 1: playing

    this._totalFrames = 0;
    this._bufferedFrames = 0;
    this._currentFrame = 0;
    this._bufferEndFrame = 0;

    this._buffer = new CircularBuffer(2 * this.BUFFER_SIZE);
    this._buffer.onChanged = this._onBufferChanged.bind(this);

    this._wsManager = new WSManager(host);

    this._timer = undefined;
}

VideoPlayer.prototype.init = function() {
    var self = this;

    Debug.Log("Begin initialization...");

    // hook up events
    this._playBtn.click(function(ev) {
    	Debug.Log("Play button clicked.");

        if (self._isPlaying) {
        	Debug.Log("Change from playing -> pause");
            self._pause();
            self._playBtn.html("Play");
        } else {
        	Debug.Log("Change from pause -> playing");
            self._play();
            self._playBtn.html("Pause");
        }
    });

    this._wsManager.RegisterHandler("GETFRAMECOUNT", function(data) {
        self._totalFrames = data.FrameCount;
        Debug.Log("Got total frame count: " + self._totalFrames);

        // start loading
        self._loadNextBlock();
    });

    this._wsManager.RegisterHandler("GETDATA", function(data) {
        self._bufferedFrames += 1;
        if (!self._buffer.Enqueue(data.Frame)) {
            Debug.Log("WARNING: buffer is full");
        }
    });

    this._wsManager.SetErrorHandler(function(data) {
        self._displayErrorMessage(data)
    });

    self._wsManager.Connect(function() {
        self._wsManager.SendCommand("GETFRAMECOUNT");
    });
};

VideoPlayer.prototype._displayErrorMessage = function(errData) {
    this._notificationElm.html("<p>Error: " + JSON.stringify(data) + "</p>");
};

VideoPlayer.prototype._onBufferChanged = function(buffer) {
    var bufferSize = buffer.Length();

    // do preloading and changing playing status according to buffer size
    if (this._playingState == 0 && (bufferSize >= this.CACHE_LIMIT || !this._hasMoreData())) {
        // cache enough frames start playing
        this._playingState = 1;
        // remove caching indication
    } else if (this._playingState == 1 && bufferSize <= Math.ceil(this.CACHE_LIMIT / 2) && this._hasMoreData()) {
        // no enough frames remaining
        this._playingState = 0;
    	// show caching indication
    }

    // load next block when there is not much frames remaining in current block
    if (this._hasMoreData() && !this._isLoading() && bufferSize <= this.FETCH_LIMIT) {
        this._loadNextBlock();
    }
};

VideoPlayer.prototype._hasMoreData = function() {
    return this._bufferEndFrame < this._totalFrames;
}

VideoPlayer.prototype._isLoading = function() {
	return this._bufferEndFrame > this._bufferedFrames;
}

VideoPlayer.prototype._loadNextBlock = function() {
    if (!this._hasMoreData()) return;

    var toFrame = Math.min(this._bufferEndFrame + this.BUFFER_SIZE, this._totalFrames);
    this._wsManager.SendCommand("GETDATA", {
        from: this._bufferEndFrame,
        to: toFrame
    });
    this._bufferEndFrame = toFrame;
};

VideoPlayer.prototype._play = function() {
    var self = this;
    this._isPlaying = 1;
    this._timer = window.setInterval(function() {
        if (self._playingState == 1) {
            var frame = self._buffer.Dequeue();
            if (frame !== undefined) {
		        var raw = atob(frame);
		        var h = pako.ungzip(raw, {
		           to: 'string'
		        });
                self._frameElm.html(h);
                self._currentFrame += 1;
            }
        }
    }, 33);
};

VideoPlayer.prototype._pause = function() {
    this._isPlaying = 0;
    if (this._timer !== undefined) {
        clearInterval(this._timer);
        this._timer = undefined;
    }
};