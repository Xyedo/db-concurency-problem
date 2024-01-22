package helper

import gonanoid "github.com/matoous/go-nanoid/v2"

func nanoid() string {
	return gonanoid.MustGenerate("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz", 16)
}

func AccountId() string {
	return "account_" + nanoid()
}

func ThreadId() string {
	return "thread_" + nanoid()
}

func CommentId() string {
	return "comment_" + nanoid()
}

func ReactionId() string {
	return "reaction_" + nanoid()
}

func FakeTableId() string {
	return "ft_" + nanoid()
}
