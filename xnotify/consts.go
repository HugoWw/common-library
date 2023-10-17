package xnotify

// Flags used as first parameter to Initiliaze func
// First flag parameter of the fanotify_init api
const (
	/* flags used for fanotify_init() */
	FAN_CLOEXEC  = 0x00000001
	FAN_NONBLOCK = 0x00000002

	/* These are NOT bitwise flags.  Both bits are used togther.  */
	FAN_CLASS_NOTIF       = 0x00000000
	FAN_CLASS_CONTENT     = 0x00000004
	FAN_CLASS_PRE_CONTENT = 0x00000008

	FAN_ALL_CLASS_BITS = FAN_CLASS_NOTIF |
		FAN_CLASS_CONTENT |
		FAN_CLASS_PRE_CONTENT

	FAN_UNLIMITED_QUEUE = 0x00000010
	FAN_UNLIMITED_MARKS = 0x00000020

	FAN_ALL_INIT_FLAGS = FAN_CLOEXEC |
		FAN_NONBLOCK |
		FAN_ALL_CLASS_BITS |
		FAN_UNLIMITED_QUEUE |
		FAN_UNLIMITED_MARKS

	/* Flags to determine fanotify event format */
	FAN_REPORT_TID     = 0x00000100 /* event->pid is thread id */
	FAN_REPORT_FID     = 0x00000200 /* Report unique file id */
	FAN_REPORT_DIR_FID = 0x00000400 /* Report unique directory id */
	FAN_REPORT_NAME    = 0x00000800 /* Report events with name */

	/* Convenience macro - FAN_REPORT_NAME requires FAN_REPORT_DIR_FID */
	FAN_REPORT_DFID_NAME = FAN_REPORT_DIR_FID | FAN_REPORT_NAME
)

// Flags used for the Mark Method
// second parameter "flags" of the fanotify_mark() api
const (
	/* flags used for fanotify_modify_mark() */
	FAN_MARK_ADD    = 0x00000001
	FAN_MARK_REMOVE = 0x00000002

	/* 如果 pathname 是符号链接，只监听符号链接而不需要监听文件本身（默认会监听文件本身） */
	FAN_MARK_DONT_FOLLOW = 0x00000004

	/* 只监听目录，如果传入的不是目录返回错误 */
	FAN_MARK_ONLYDIR             = 0x00000008
	FAN_MARK_MOUNT               = 0x00000010
	FAN_MARK_IGNORED_MASK        = 0x00000020
	FAN_MARK_IGNORED_SURV_MODIFY = 0x00000040

	/* 移除所有 marks */
	FAN_MARK_FLUSH = 0x00000080

	FAN_ALL_MARK_FLAGS = FAN_MARK_ADD |
		FAN_MARK_REMOVE |
		FAN_MARK_DONT_FOLLOW |
		FAN_MARK_ONLYDIR |
		FAN_MARK_MOUNT |
		FAN_MARK_IGNORED_MASK |
		FAN_MARK_IGNORED_SURV_MODIFY |
		FAN_MARK_FLUSH
)

// Event types
// third parameter "mask" of the fanotify_mark() api
const (
	FAN_ACCESS        = 0x00000001 /* File was accessed */
	FAN_MODIFY        = 0x00000002 /* File was modified */
	FAN_ATTRIB        = 0x00000004 /* Metadata changed */
	FAN_CLOSE_WRITE   = 0x00000008 /* Writtable file closed */
	FAN_CLOSE_NOWRITE = 0x00000010 /* Unwrittable file closed */
	FAN_OPEN          = 0x00000020 /* File was opened */

	FAN_MOVED_FROM = 0x00000040 /* File was moved from X */
	FAN_MOVED_TO   = 0x00000080 /* File was moved to Y */

	FAN_CREATE      = 0x00000100 /* Subfile was created */
	FAN_DELETE      = 0x00000200 /* Subfile was deleted */
	FAN_DELETE_SELF = 0x00000400 /* Self was deleted */
	FAN_MOVE_SELF   = 0x00000800 /* Self was moved */
	FAN_OPEN_EXEC   = 0x00001000 /* File was opened for exec */

	FAN_Q_OVERFLOW = 0x00004000 /* Event queued overflowed */
	FAN_FS_ERROR   = 0x00008000 /* Filesystem error */

	FAN_OPEN_PERM      = 0x00010000 /* File open in perm check */
	FAN_ACCESS_PERM    = 0x00020000 /* File accessed in perm check */
	FAN_OPEN_EXEC_PERM = 0x00040000 /* File open/exec in perm check */

	FAN_RENAME = 0x10000000 /* File was renamed */
	FAN_ONDIR  = 0x40000000 /* event occurred against dir */

	FAN_EVENT_ON_CHILD = 0x08000000 /* interested in child events */

	/* helper events */
	FAN_CLOSE = FAN_CLOSE_WRITE | FAN_CLOSE_NOWRITE /* close */

	/*
	 * All of the events - we build the list by hand so that we can add flags in
	 * the future and not break backward compatibility.  Apps will get only the
	 * events that they originally wanted.  Be sure to add new events here!
	 */
	FAN_ALL_EVENTS = FAN_ACCESS |
		FAN_MODIFY |
		FAN_CLOSE |
		FAN_OPEN

	/*
	 * All events which require a permission response from userspace
	 */
	FAN_ALL_PERM_EVENTS = FAN_OPEN_PERM |
		FAN_ACCESS_PERM

	FAN_ALL_OUTGOING_EVENTS = FAN_ALL_EVENTS |
		FAN_ALL_PERM_EVENTS |
		FAN_Q_OVERFLOW

	FANOTIFY_METADATA_VERSION = 3

	FAN_ALLOW = 0x01
	FAN_DENY  = 0x02
	FAN_NOFD  = -1
)

const FA_EVENT_LEN = 24
const maxStatCmdLen = 15
