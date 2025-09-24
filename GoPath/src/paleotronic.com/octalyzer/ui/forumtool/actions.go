package forumtool

type ForumAction int

const (
	forumQuitAll ForumAction = iota
	forumContinue
	forumMarkRead
	forumNextMessage
	forumReturn
	forumPrevMessage
	forumReplyMessage
	forumViewReplies
	forumViewParent
	forumLikeMessage
	forumShowParentMessage
	forumNextForum
	forumPrevForum
	forumMarkAllInForumRead
	forumMarkAllForumsRead
	forumQuitScan
)
