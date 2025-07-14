// retrieving data from a form
/*
 * a form containing:
 * - email
 * - password
 */

(def email
     (hget body %email))

(def password
     (hget body %password))

(def message
     (concat
       "your email is" email
       ", and your password is " password))

(hash message: message)
