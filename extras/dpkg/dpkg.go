package dpkg

// #cgo LDFLAGS: -ldpkg -lmd
// #define LIBDPKG_VOLATILE_API 1
//
// #include <dpkg/dpkg.h>
// #include <dpkg/dpkg-db.h>
// #include <dpkg/pkg-array.h>
import "C"
import (
	"errors"
	"strconv"
)

var DpkgInstanced bool

func NewDpkgInstance() error {
	if DpkgInstanced {
		return errors.New("another dpkg instance already exists")
	}

	C.dpkg_program_init(C.CString("a.out"))
	C.modstatdb_open(C.msdbrw_available_readonly)

	DpkgInstanced = true
	return nil
}

func DpkgDispose() error {
	if !DpkgInstanced {
		return errors.New("no dpkg instance exists")
	}

	C.dpkg_program_done()

	DpkgInstanced = false
	return nil
}

func getPackageVersion(pkgName string) string {
	pkgInfo := C.pkg_hash_find_singleton(C.CString(pkgName))

	versionEpoch := int(pkgInfo.configversion.epoch)
	version := C.GoString(pkgInfo.configversion.version)
	versionRevision := C.GoString(pkgInfo.configversion.revision)

	versionStr := ""
	if versionEpoch != 0 {
		versionStr += strconv.Itoa(versionEpoch) + ":"
	}
	versionStr += version
	if versionRevision != "" {
		versionStr += "-" + versionRevision
	}

	return versionStr
}

func DpkgGetPackageVersion(pkgName string) string {
	NewDpkgInstance()
	version := getPackageVersion(pkgName)
	DpkgDispose()

	return version
}

func DpkgBatchGetPackageVersion(pkgNames []string) []string {
	versions := make([]string, len(pkgNames))

	NewDpkgInstance()
	for i := 0; i < len(pkgNames); i++ {
		versions[i] = getPackageVersion(pkgNames[i])
	}
	DpkgDispose()

	return versions
}
