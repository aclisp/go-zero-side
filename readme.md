# Go zero side

Go-zero-side is a project to complement go-zero. It provides *more practical engineering tools* which work seamlessly with go-zero.

The packages are:

* `goschedule` A distributed cron job scheduler. It has a simple design to prevent multiple nodes from running the same job without using distributed locks.

* `logz` An xorm logger implement to show SQL by integrating with go-zero logging facility. It also outputs trace-id for SQL logs.

* `xormcache` *Deprecated.* An xorm cache implement using redis as the storage. It is deprecated because a remote cache is not so efficient as xorm's build-in local cache such as memory or leveldb.

* `session` A HTTP session middleware, based on gorilla/sessions but far better than that.

* `router` A customied go-zero/rest/httpx.Router. Currently it creates a fileServingRouter which handles HTTP caching correctly.

* `embedx` handles go template files embedding as well as static files embedding.

* `delayq` defines a general purpose Message struct as the protocol of zeromicro/go-queue/dq. It also propagates trace-id to the queue consumer.
