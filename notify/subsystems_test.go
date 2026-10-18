package notify

// The tests start an app of notify alone, whose mailer needs the framework's
// mail subsystem: the binary links it only by this import, as main.go's does
// for the product.
import _ "github.com/kitsunium/sdk/framework/app/mail"
