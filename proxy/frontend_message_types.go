package proxy

// case 'B':
// 		msg = &b.bind
// 	case 'C':
// 		msg = &b._close
// 	case 'D':
// 		msg = &b.describe
// 	case 'E':
// 		msg = &b.execute
// 	case 'f':
// 		msg = &b.copyFail
// 	case 'H':
// 		msg = &b.flush
// 	case 'P':
// 		msg = &b.parse
// 	case 'p':
// 		msg = &b.passwordMessage
// 	case 'Q':
// 		msg = &b.query
// 	case 'S':
// 		msg = &b.sync
// 	case 'X':
// 		msg = &b.terminate

type QueryMessage struct {
	Type   string
	String string
}
