(cond init
  (broadcast "general" (concat conn_id ": " msg))
  (subscribe "general" conn_id))
