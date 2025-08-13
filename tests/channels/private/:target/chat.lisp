(cond init
  (broadcast (hget params %target) (concat conn_id ": " msg))
  (subscribe (hget params %target) conn_id))
