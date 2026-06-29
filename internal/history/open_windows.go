//go:build windows

package history

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// openAppend opens the history file for appending, creating it with an owner-only
// security descriptor if absent — the Windows equivalent of Unix 0600. The DACL
// is protected (no inheritance from the parent directory) and grants full file
// access to the current user alone, so secrets in variable bindings are not
// readable by other principals on a shared host.
//
// The security descriptor is applied only when the file is created; on an
// existing file CreateFile ignores it, so the owner-only DACL set at creation
// time persists across later appends.
func openAppend(path string) (*os.File, error) {
	sd, err := ownerOnlyDescriptor()
	if err != nil {
		return nil, err
	}
	sa := &windows.SecurityAttributes{SecurityDescriptor: sd}
	sa.Length = uint32(unsafe.Sizeof(*sa))

	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("history: path: %w", err)
	}
	h, err := windows.CreateFile(
		p,
		windows.FILE_APPEND_DATA,
		windows.FILE_SHARE_READ,
		sa,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("history: create: %w", err)
	}
	return os.NewFile(uintptr(h), path), nil
}

// ownerOnlyDescriptor builds a security descriptor whose owner is the current
// user and whose protected DACL contains a single ACE granting that user full
// file access — no other principal, no inherited ACEs.
func ownerOnlyDescriptor() (*windows.SECURITY_DESCRIPTOR, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, fmt.Errorf("history: token user: %w", err)
	}
	sid := user.User.Sid.String()
	// SDDL: owner+group = current user; D:P = protected DACL (blocks inheritance);
	// one ACE (A) granting FILE_ALL_ACCESS (FA) to the user SID. No other ACEs.
	sddl := fmt.Sprintf("O:%[1]sG:%[1]sD:P(A;;FA;;;%[1]s)", sid)
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return nil, fmt.Errorf("history: build security descriptor: %w", err)
	}
	return sd, nil
}
