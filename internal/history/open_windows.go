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
	// Whether the file already exists decides how it is secured. A new file gets its
	// owner-only DACL atomically from SecurityAttributes at creation — privilege-free
	// and guaranteed. A pre-existing file is additionally re-tightened, but only
	// best-effort: rewriting an existing file's ACL can be denied in some
	// environments, and a file nock created is already owner-only, so a failed
	// re-tighten must never block an append.
	preexisting := false
	if _, statErr := os.Stat(path); statErr == nil {
		preexisting = true
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

	if preexisting {
		// Defense-in-depth for a file that existed before this open (e.g. created
		// outside nock with a looser ACL). Ignore a denial: nock-created files are
		// already owner-only from their SecurityAttributes at creation time.
		_ = enforceOwnerOnly(h, sd)
	}
	return os.NewFile(uintptr(h), path), nil
}

// enforceOwnerOnly reapplies the protected, owner-only DACL from sd onto an
// already-open handle, correcting a pre-existing file's permissions rather than
// inheriting them. Only the DACL is set: the creating user is already the owner,
// and rewriting the owner at runtime needs SeRestorePrivilege (denied to ordinary
// processes, e.g. CI runners). The DACL is what actually gates access. Callers
// invoke this best-effort — new files are already secured at creation, so a
// failure here does not compromise a file nock itself wrote.
func enforceOwnerOnly(h windows.Handle, sd *windows.SECURITY_DESCRIPTOR) error {
	dacl, _, err := sd.DACL()
	if err != nil {
		return fmt.Errorf("history: read dacl: %w", err)
	}
	info := windows.SECURITY_INFORMATION(
		windows.DACL_SECURITY_INFORMATION |
			windows.PROTECTED_DACL_SECURITY_INFORMATION,
	)
	if err := windows.SetSecurityInfo(h, windows.SE_FILE_OBJECT, info, nil, nil, dacl, nil); err != nil {
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
