//go:build windows

package history

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// openAppend opens the history file for appending, enforcing an owner-only
// security descriptor — the Windows equivalent of Unix 0600. The descriptor's
// DACL is protected (no inheritance from the parent directory) and grants full
// file access to the current user alone, so secrets in variable bindings are not
// readable by other principals on a shared host.
//
// The descriptor is set two ways so both new and existing files are covered:
//   - SecurityAttributes applies it atomically at creation time.
//   - SetSecurityInfo reapplies it on every open, so a file created or migrated
//     with a looser ACL is re-tightened before it receives secrets (CreateFile's
//     OPEN_ALWAYS otherwise reuses the existing ACL untouched).
//
// FILE_SHARE_READ lets a concurrent reader (e.g. Recent) open the file while an
// append is in flight; access is still governed by the owner-only DACL, not the
// share mode.
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
		windows.FILE_APPEND_DATA|windows.WRITE_DAC,
		windows.FILE_SHARE_READ,
		sa,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("history: create: %w", err)
	}

	if err := enforceOwnerOnly(h, sd); err != nil {
		_ = windows.CloseHandle(h)
		return nil, err
	}
	return os.NewFile(uintptr(h), path), nil
}

// enforceOwnerOnly reapplies the owner, group, and protected DACL from sd onto an
// already-open handle, so an existing file's permissions are corrected rather
// than inherited. Failing here is fail-closed: the caller discards the handle.
func enforceOwnerOnly(h windows.Handle, sd *windows.SECURITY_DESCRIPTOR) error {
	owner, _, err := sd.Owner()
	if err != nil {
		return fmt.Errorf("history: read owner: %w", err)
	}
	group, _, err := sd.Group()
	if err != nil {
		return fmt.Errorf("history: read group: %w", err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return fmt.Errorf("history: read dacl: %w", err)
	}
	info := windows.SECURITY_INFORMATION(
		windows.OWNER_SECURITY_INFORMATION |
			windows.GROUP_SECURITY_INFORMATION |
			windows.DACL_SECURITY_INFORMATION |
			windows.PROTECTED_DACL_SECURITY_INFORMATION,
	)
	if err := windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, info, owner, group, dacl, nil); err != nil {
		return fmt.Errorf("history: enforce owner-only acl: %w", err)
	}
	return nil
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
