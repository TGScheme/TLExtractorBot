package types

type PatchOS string

const (
	AndroidPatch  PatchOS = "android"
	IOSPatch      PatchOS = "ios"
	TDesktopPatch PatchOS = "tdesktop"
	TDLibPatch    PatchOS = "tdlib"
)
