package persona

// isOwnerDirectChat is true only for the stored 1:1 with the persona owner.
func isOwnerDirectChat(ownerChatID, dcChatID uint32) bool {
	return ownerChatID != 0 && dcChatID == ownerChatID
}
