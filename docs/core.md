

# core
`import "github.com/vanilla-os/abroot/core"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)

## <a name="pkg-overview">Overview</a>



## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func AtomicRsync(src, dst string, transitionalPath string, finalPath string, excluded []string, keepUnwanted bool) error](#AtomicRsync)
* [func AtomicSwap(src, dst string) error](#AtomicSwap)
* [func BaseImagePackageDiff(currentDigest, newDigest string) (added, upgraded, downgraded, removed []diff.PackageDiff, err error)](#BaseImagePackageDiff)
* [func CopyFile(source, dest string) error](#CopyFile)
* [func DeleteImageForRoot(root string) error](#DeleteImageForRoot)
* [func DiffFiles(sourceFile, destFile string) ([]byte, error)](#DiffFiles)
* [func FindImageWithLabel(key, value string) (string, error)](#FindImageWithLabel)
* [func GetLogFile() *os.File](#GetLogFile)
* [func GetRepoContentsForPkg(pkg string) (map[string]interface{}, error)](#GetRepoContentsForPkg)
* [func GetToken() (string, error)](#GetToken)
* [func IsVerbose() bool](#IsVerbose)
* [func KargsBackup() error](#KargsBackup)
* [func KargsEdit() (bool, error)](#KargsEdit)
* [func KargsFormat(content string) (string, error)](#KargsFormat)
* [func KargsRead() (string, error)](#KargsRead)
* [func KargsWrite(content string) error](#KargsWrite)
* [func LogToFile(msg string, args ...interface{}) error](#LogToFile)
* [func MergeDiff(firstFile, secondFile, destination string) error](#MergeDiff)
* [func OciExportRootFs(buildImageName string, imageRecipe *ImageRecipe, transDir string, dest string) error](#OciExportRootFs)
* [func OverlayPackageDiff() (added, upgraded, downgraded, removed []diff.PackageDiff, err error)](#OverlayPackageDiff)
* [func PrintVerbose(prefix, level string, depth float32, args ...interface{})](#PrintVerbose)
* [func PrintVerboseErr(prefix string, depth float32, args ...interface{})](#PrintVerboseErr)
* [func PrintVerboseErrNoLog(prefix string, depth float32, args ...interface{})](#PrintVerboseErrNoLog)
* [func PrintVerboseInfo(prefix string, args ...interface{})](#PrintVerboseInfo)
* [func PrintVerboseInfoNoLog(prefix string, args ...interface{})](#PrintVerboseInfoNoLog)
* [func PrintVerboseNoLog(prefix, level string, depth float32, args ...interface{})](#PrintVerboseNoLog)
* [func PrintVerboseSimple(args ...interface{})](#PrintVerboseSimple)
* [func PrintVerboseSimpleNoLog(args ...interface{})](#PrintVerboseSimpleNoLog)
* [func PrintVerboseWarn(prefix string, depth float32, args ...interface{})](#PrintVerboseWarn)
* [func PrintVerboseWarnNoLog(prefix string, depth float32, args ...interface{})](#PrintVerboseWarnNoLog)
* [func RetrieveImageForRoot(root string) (string, error)](#RetrieveImageForRoot)
* [func RootCheck(display bool) bool](#RootCheck)
* [func WriteDiff(destFile string, diffLines []byte) error](#WriteDiff)
* [type ABImage](#ABImage)
  * [func NewABImage(digest string, image string) (*ABImage, error)](#NewABImage)
  * [func NewABImageFromRoot() (*ABImage, error)](#NewABImageFromRoot)
  * [func (a *ABImage) WriteTo(dest string, suffix string) error](#ABImage.WriteTo)
* [type ABRollbackResponse](#ABRollbackResponse)
* [type ABRootManager](#ABRootManager)
  * [func NewABRootManager() *ABRootManager](#NewABRootManager)
  * [func (a *ABRootManager) GetBoot() (partition Partition, err error)](#ABRootManager.GetBoot)
  * [func (a *ABRootManager) GetFuture() (partition ABRootPartition, err error)](#ABRootManager.GetFuture)
  * [func (a *ABRootManager) GetInit() (partition Partition, err error)](#ABRootManager.GetInit)
  * [func (a *ABRootManager) GetOther() (partition ABRootPartition, err error)](#ABRootManager.GetOther)
  * [func (a *ABRootManager) GetPartition(label string) (partition ABRootPartition, err error)](#ABRootManager.GetPartition)
  * [func (a *ABRootManager) GetPartitions() error](#ABRootManager.GetPartitions)
  * [func (a *ABRootManager) GetPresent() (partition ABRootPartition, err error)](#ABRootManager.GetPresent)
  * [func (a *ABRootManager) IdentifyPartition(partition Partition) (identifiedAs string, err error)](#ABRootManager.IdentifyPartition)
  * [func (a *ABRootManager) IsCurrent(partition Partition) bool](#ABRootManager.IsCurrent)
* [type ABRootPartition](#ABRootPartition)
* [type ABSystem](#ABSystem)
  * [func NewABSystem() (*ABSystem, error)](#NewABSystem)
  * [func (s *ABSystem) AddToCleanUpQueue(name string, priority int, values ...interface{})](#ABSystem.AddToCleanUpQueue)
  * [func (s *ABSystem) CheckAll() error](#ABSystem.CheckAll)
  * [func (s *ABSystem) CheckUpdate() (string, bool)](#ABSystem.CheckUpdate)
  * [func (s *ABSystem) CreateStageFile() error](#ABSystem.CreateStageFile)
  * [func (s *ABSystem) GenerateCrypttab(rootPath string) error](#ABSystem.GenerateCrypttab)
  * [func (s *ABSystem) GenerateFstab(rootPath string, root ABRootPartition) error](#ABSystem.GenerateFstab)
  * [func (s *ABSystem) GenerateSystemdUnits(rootPath string, root ABRootPartition) error](#ABSystem.GenerateSystemdUnits)
  * [func (s *ABSystem) LockUpgrade() error](#ABSystem.LockUpgrade)
  * [func (s *ABSystem) RemoveFromCleanUpQueue(name string)](#ABSystem.RemoveFromCleanUpQueue)
  * [func (s *ABSystem) RemoveStageFile() error](#ABSystem.RemoveStageFile)
  * [func (s *ABSystem) ResetQueue()](#ABSystem.ResetQueue)
  * [func (s *ABSystem) Rollback() (response ABRollbackResponse, err error)](#ABSystem.Rollback)
  * [func (s *ABSystem) RunCleanUpQueue(fnName string) error](#ABSystem.RunCleanUpQueue)
  * [func (s *ABSystem) RunOperation(operation ABSystemOperation) error](#ABSystem.RunOperation)
  * [func (s *ABSystem) UnlockUpgrade() error](#ABSystem.UnlockUpgrade)
  * [func (s *ABSystem) UpgradeLockExists() bool](#ABSystem.UpgradeLockExists)
  * [func (s *ABSystem) UserLockRequested() bool](#ABSystem.UserLockRequested)
* [type ABSystemOperation](#ABSystemOperation)
* [type Checks](#Checks)
  * [func NewChecks() *Checks](#NewChecks)
  * [func (c *Checks) CheckCompatibilityFS() error](#Checks.CheckCompatibilityFS)
  * [func (c *Checks) CheckConnectivity() error](#Checks.CheckConnectivity)
  * [func (c *Checks) CheckRoot() error](#Checks.CheckRoot)
  * [func (c *Checks) PerformAllChecks() error](#Checks.PerformAllChecks)
* [type Children](#Children)
* [type Chroot](#Chroot)
  * [func NewChroot(root string, rootUuid string, rootDevice string) (*Chroot, error)](#NewChroot)
  * [func (c *Chroot) Close() error](#Chroot.Close)
  * [func (c *Chroot) Execute(cmd string, args []string) error](#Chroot.Execute)
  * [func (c *Chroot) ExecuteCmds(cmds []string) error](#Chroot.ExecuteCmds)
* [type DiskManager](#DiskManager)
  * [func NewDiskManager() *DiskManager](#NewDiskManager)
  * [func (d *DiskManager) GetPartitionByLabel(label string) (Partition, error)](#DiskManager.GetPartitionByLabel)
* [type GPUInfo](#GPUInfo)
* [type Grub](#Grub)
  * [func NewGrub(bootPart Partition) (*Grub, error)](#NewGrub)
  * [func (g *Grub) IsBootedIntoPresentRoot() (bool, error)](#Grub.IsBootedIntoPresentRoot)
* [type ImageRecipe](#ImageRecipe)
  * [func NewImageRecipe(image string, labels map[string]string, args map[string]string, content string) *ImageRecipe](#NewImageRecipe)
  * [func (c *ImageRecipe) Write(path string) error](#ImageRecipe.Write)
* [type IntegrityCheck](#IntegrityCheck)
  * [func NewIntegrityCheck(root ABRootPartition, repair bool) (*IntegrityCheck, error)](#NewIntegrityCheck)
  * [func (ic *IntegrityCheck) Repair() error](#IntegrityCheck.Repair)
* [type Manifest](#Manifest)
* [type PCSpecs](#PCSpecs)
  * [func GetPCSpecs() PCSpecs](#GetPCSpecs)
* [type PackageManager](#PackageManager)
  * [func NewPackageManager(dryRun bool) *PackageManager](#NewPackageManager)
  * [func (p *PackageManager) Add(pkg string) error](#PackageManager.Add)
  * [func (p *PackageManager) ClearUnstagedPackages() error](#PackageManager.ClearUnstagedPackages)
  * [func (p *PackageManager) ExistsInRepo(pkg string) error](#PackageManager.ExistsInRepo)
  * [func (p *PackageManager) GetAddPackages() ([]string, error)](#PackageManager.GetAddPackages)
  * [func (p *PackageManager) GetAddPackagesString(sep string) (string, error)](#PackageManager.GetAddPackagesString)
  * [func (p *PackageManager) GetFinalCmd(operation ABSystemOperation) string](#PackageManager.GetFinalCmd)
  * [func (p *PackageManager) GetRemovePackages() ([]string, error)](#PackageManager.GetRemovePackages)
  * [func (p *PackageManager) GetRemovePackagesString(sep string) (string, error)](#PackageManager.GetRemovePackagesString)
  * [func (p *PackageManager) GetUnstagedPackages() ([]UnstagedPackage, error)](#PackageManager.GetUnstagedPackages)
  * [func (p *PackageManager) GetUnstagedPackagesPlain() ([]string, error)](#PackageManager.GetUnstagedPackagesPlain)
  * [func (p *PackageManager) Remove(pkg string) error](#PackageManager.Remove)
* [type Partition](#Partition)
  * [func (p *Partition) IsDevMapper() bool](#Partition.IsDevMapper)
  * [func (p *Partition) IsEncrypted() bool](#Partition.IsEncrypted)
  * [func (p *Partition) Mount(destination string) error](#Partition.Mount)
  * [func (p *Partition) Unmount() error](#Partition.Unmount)
* [type QueuedFunction](#QueuedFunction)
* [type Registry](#Registry)
  * [func NewRegistry() *Registry](#NewRegistry)
  * [func (r *Registry) GetManifest(token string) (*Manifest, error)](#Registry.GetManifest)
  * [func (r *Registry) HasUpdate(digest string) (string, bool)](#Registry.HasUpdate)
* [type UnstagedPackage](#UnstagedPackage)


#### <a name="pkg-files">Package files</a>
[atomic-io.go](/src/github.com/vanilla-os/abroot/core/atomic-io.go) [checks.go](/src/github.com/vanilla-os/abroot/core/checks.go) [chroot.go](/src/github.com/vanilla-os/abroot/core/chroot.go) [diff.go](/src/github.com/vanilla-os/abroot/core/diff.go) [disk-manager.go](/src/github.com/vanilla-os/abroot/core/disk-manager.go) [grub.go](/src/github.com/vanilla-os/abroot/core/grub.go) [image-recipe.go](/src/github.com/vanilla-os/abroot/core/image-recipe.go) [image.go](/src/github.com/vanilla-os/abroot/core/image.go) [integrity.go](/src/github.com/vanilla-os/abroot/core/integrity.go) [kargs.go](/src/github.com/vanilla-os/abroot/core/kargs.go) [logging.go](/src/github.com/vanilla-os/abroot/core/logging.go) [oci.go](/src/github.com/vanilla-os/abroot/core/oci.go) [package-diff.go](/src/github.com/vanilla-os/abroot/core/package-diff.go) [packages.go](/src/github.com/vanilla-os/abroot/core/packages.go) [registry.go](/src/github.com/vanilla-os/abroot/core/registry.go) [root.go](/src/github.com/vanilla-os/abroot/core/root.go) [rsync.go](/src/github.com/vanilla-os/abroot/core/rsync.go) [specs.go](/src/github.com/vanilla-os/abroot/core/specs.go) [system.go](/src/github.com/vanilla-os/abroot/core/system.go) [utils.go](/src/github.com/vanilla-os/abroot/core/utils.go) 


## <a name="pkg-constants">Constants</a>
``` go
const (
    DefaultKargs = "quiet splash bgrt_disable $vt_handoff"
    KargsTmpFile = "/tmp/kargs-temp"
)
```
``` go
const (
    PackagesBaseDir       = "/etc/abroot"
    DryRunPackagesBaseDir = "/tmp/abroot"
    PackagesAddFile       = "packages.add"
    PackagesRemoveFile    = "packages.remove"
    PackagesUnstagedFile  = "packages.unstaged"
)
```
``` go
const (
    ADD    = "+"
    REMOVE = "-"
)
```
``` go
const (
    // ABSystem operations
    UPGRADE           = "upgrade"
    FORCE_UPGRADE     = "force-upgrade"
    DRY_RUN_UPGRADE   = "dry-run-upgrade"
    APPLY             = "package-apply"
    DRY_RUN_APPLY     = "dry-run-package-apply"
    INITRAMFS         = "initramfs"
    DRY_RUN_INITRAMFS = "dry-run-initramfs"

    // ABSystem rollback response
    ROLLBACK_UNNECESSARY = "rollback-unnecessary"
    ROLLBACK_SUCCESS     = "rollback-success"
    ROLLBACK_FAILED      = "rollback-failed"
)
```
Supported ABSystemOperation types


## <a name="pkg-variables">Variables</a>
``` go
var (

    // Errors
    ErrNoUpdate error = errors.New("no update available")
)
```
Common variables and errors used by the ABSystem

``` go
var KargsPath = "/etc/abroot/kargs"
```
``` go
var ReservedMounts = []string{
    "/dev",
    "/dev/pts",
    "/proc",
    "/run",
    "/sys",
}
```
ReservedMounts is a list of mount points from host which should be
mounted inside the chroot environment to ensure it works properly in
some cases, such as grub-mkconfig



## <a name="AtomicRsync">func</a> [AtomicRsync](/src/target/rsync.go?s=3049:3169#L120)
``` go
func AtomicRsync(src, dst string, transitionalPath string, finalPath string, excluded []string, keepUnwanted bool) error
```
AtomicRsync executes the rsync command in an atomic-like manner.
It does so by dry-running the rsync, and if it succeeds, it runs
the rsync again performing changes.
If the keepUnwanted option
is set to true, it will omit the --delete option, so that the already
existing and unwanted files will not be deleted.
To ensure the changes are applied atomically, we rsync on a _new directory first,
and use atomicSwap to replace the _new with the dst directory.



## <a name="AtomicSwap">func</a> [AtomicSwap](/src/target/atomic-io.go?s=685:723#L26)
``` go
func AtomicSwap(src, dst string) error
```
atomicSwap allows swapping 2 files or directories in-place and atomically,
using the renameat2 syscall. This should be used instead of os.Rename,
which is not atomic at all



## <a name="BaseImagePackageDiff">func</a> [BaseImagePackageDiff](/src/target/package-diff.go?s=732:864#L30)
``` go
func BaseImagePackageDiff(currentDigest, newDigest string) (
    added, upgraded, downgraded, removed []diff.PackageDiff,
    err error,
)
```
BaseImagePackageDiff retrieves the added, removed, upgraded and downgraded
base packages (the ones bundled with the image).



## <a name="CopyFile">func</a> [CopyFile](/src/target/utils.go?s=1432:1472#L74)
``` go
func CopyFile(source, dest string) error
```
CopyFile copies a file from source to dest



## <a name="DeleteImageForRoot">func</a> [DeleteImageForRoot](/src/target/oci.go?s=4078:4120#L166)
``` go
func DeleteImageForRoot(root string) error
```
DeleteImageForRoot deletes the image created for the provided root



## <a name="DiffFiles">func</a> [DiffFiles](/src/target/diff.go?s=1621:1680#L60)
``` go
func DiffFiles(sourceFile, destFile string) ([]byte, error)
```
DiffFiles returns the diff lines between source and dest files using the
diff command (assuming it is installed). If no diff is found, nil is
returned. If any errors occur, they are returned.



## <a name="FindImageWithLabel">func</a> [FindImageWithLabel](/src/target/oci.go?s=2664:2722#L114)
``` go
func FindImageWithLabel(key, value string) (string, error)
```
FindImageWithLabel returns the name of the first image containing the
provided key-value pair or an empty string if none was found



## <a name="GetLogFile">func</a> [GetLogFile](/src/target/logging.go?s=5133:5159#L173)
``` go
func GetLogFile() *os.File
```
GetLogFile returns the log file handle



## <a name="GetRepoContentsForPkg">func</a> [GetRepoContentsForPkg](/src/target/packages.go?s=14219:14289#L520)
``` go
func GetRepoContentsForPkg(pkg string) (map[string]interface{}, error)
```
GetRepoContentsForPkg retrieves package information from the repository API



## <a name="GetToken">func</a> [GetToken](/src/target/registry.go?s=2123:2154#L75)
``` go
func GetToken() (string, error)
```
GetToken generates a token using the provided tokenURL and returns it



## <a name="IsVerbose">func</a> [IsVerbose](/src/target/logging.go?s=1844:1865#L73)
``` go
func IsVerbose() bool
```
IsVerbose checks if verbose mode is enabled



## <a name="KargsBackup">func</a> [KargsBackup](/src/target/kargs.go?s=1938:1962#L86)
``` go
func KargsBackup() error
```
KargsBackup makes a backup of the current kargs file



## <a name="KargsEdit">func</a> [KargsEdit](/src/target/kargs.go?s=3799:3829#L164)
``` go
func KargsEdit() (bool, error)
```
KargsEdit copies the kargs file to a temporary file and opens it in the
user's preferred editor by querying the $EDITOR environment variable.
Once closed, its contents are written back to the main kargs file.
This function returns a boolean parameter indicating whether any changes
were made to the kargs file.



## <a name="KargsFormat">func</a> [KargsFormat](/src/target/kargs.go?s=2875:2923#L127)
``` go
func KargsFormat(content string) (string, error)
```
KargsFormat formats the contents of the kargs file, ensuring that
there are no duplicate entries, multiple spaces or trailing newline



## <a name="KargsRead">func</a> [KargsRead](/src/target/kargs.go?s=2358:2390#L106)
``` go
func KargsRead() (string, error)
```
KargsRead reads the content of the kargs file



## <a name="KargsWrite">func</a> [KargsWrite](/src/target/kargs.go?s=1296:1333#L54)
``` go
func KargsWrite(content string) error
```
KargsWrite makes a backup of the current kargs file and then
writes the new content to it



## <a name="LogToFile">func</a> [LogToFile](/src/target/logging.go?s=4850:4903#L159)
``` go
func LogToFile(msg string, args ...interface{}) error
```
LogToFile writes messages to the log file



## <a name="MergeDiff">func</a> [MergeDiff](/src/target/diff.go?s=592:655#L24)
``` go
func MergeDiff(firstFile, secondFile, destination string) error
```
MergeDiff merges the diff lines between the first and second files into
the destination file. If any errors occur, they are returned.



## <a name="OciExportRootFs">func</a> [OciExportRootFs](/src/target/oci.go?s=692:797#L29)
``` go
func OciExportRootFs(buildImageName string, imageRecipe *ImageRecipe, transDir string, dest string) error
```
OciExportRootFs generates a rootfs from an image recipe file



## <a name="OverlayPackageDiff">func</a> [OverlayPackageDiff](/src/target/package-diff.go?s=2487:2586#L87)
``` go
func OverlayPackageDiff() (
    added, upgraded, downgraded, removed []diff.PackageDiff,
    err error,
)
```
OverlayPackageDiff retrieves the added, removed, upgraded and downgraded
overlay packages (the ones added manually via `abroot pkg add`).



## <a name="PrintVerbose">func</a> [PrintVerbose](/src/target/logging.go?s=3027:3102#L112)
``` go
func PrintVerbose(prefix, level string, depth float32, args ...interface{})
```
PrintVerbose prints verbose messages and logs to the file if enabled



## <a name="PrintVerboseErr">func</a> [PrintVerboseErr](/src/target/logging.go?s=3871:3942#L134)
``` go
func PrintVerboseErr(prefix string, depth float32, args ...interface{})
```
PrintVerboseErr prints verbose error messages and logs to the file if enabled



## <a name="PrintVerboseErrNoLog">func</a> [PrintVerboseErrNoLog](/src/target/logging.go?s=3658:3734#L129)
``` go
func PrintVerboseErrNoLog(prefix string, depth float32, args ...interface{})
```
PrintVerboseErrNoLog prints verbose error messages without logging to the file



## <a name="PrintVerboseInfo">func</a> [PrintVerboseInfo](/src/target/logging.go?s=4699:4756#L154)
``` go
func PrintVerboseInfo(prefix string, args ...interface{})
```
PrintVerboseInfo prints verbose info messages and logs to the file if enabled



## <a name="PrintVerboseInfoNoLog">func</a> [PrintVerboseInfoNoLog](/src/target/logging.go?s=4502:4564#L149)
``` go
func PrintVerboseInfoNoLog(prefix string, args ...interface{})
```
PrintVerboseInfoNoLog prints verbose info messages without logging to the file



## <a name="PrintVerboseNoLog">func</a> [PrintVerboseNoLog](/src/target/logging.go?s=2747:2827#L104)
``` go
func PrintVerboseNoLog(prefix, level string, depth float32, args ...interface{})
```
PrintVerboseNoLog prints verbose messages without logging to the file



## <a name="PrintVerboseSimple">func</a> [PrintVerboseSimple](/src/target/logging.go?s=3491:3535#L124)
``` go
func PrintVerboseSimple(args ...interface{})
```
PrintVerboseSimple prints simple verbose messages and logs to the file if enabled



## <a name="PrintVerboseSimpleNoLog">func</a> [PrintVerboseSimpleNoLog](/src/target/logging.go?s=3311:3360#L119)
``` go
func PrintVerboseSimpleNoLog(args ...interface{})
```
PrintVerboseSimpleNoLog prints simple verbose messages without logging to the file



## <a name="PrintVerboseWarn">func</a> [PrintVerboseWarn](/src/target/logging.go?s=4296:4368#L144)
``` go
func PrintVerboseWarn(prefix string, depth float32, args ...interface{})
```
PrintVerboseWarn prints verbose warning messages and logs to the file if enabled



## <a name="PrintVerboseWarnNoLog">func</a> [PrintVerboseWarnNoLog](/src/target/logging.go?s=4078:4155#L139)
``` go
func PrintVerboseWarnNoLog(prefix string, depth float32, args ...interface{})
```
PrintVerboseWarnNoLog prints verbose warning messages without logging to the file



## <a name="RetrieveImageForRoot">func</a> [RetrieveImageForRoot](/src/target/oci.go?s=3729:3783#L153)
``` go
func RetrieveImageForRoot(root string) (string, error)
```
RetrieveImageForRoot retrieves the image created for the provided root
based on the label. Note for distro maintainers: labels must follow those
defined in the ABRoot config file



## <a name="RootCheck">func</a> [RootCheck](/src/target/utils.go?s=690:723#L39)
``` go
func RootCheck(display bool) bool
```


## <a name="WriteDiff">func</a> [WriteDiff](/src/target/diff.go?s=2356:2411#L86)
``` go
func WriteDiff(destFile string, diffLines []byte) error
```
WriteDiff applies the diff lines to the destination file using the patch
command (assuming it is installed). If any errors occur, they are returned.




## <a name="ABImage">type</a> [ABImage](/src/target/image.go?s=717:853#L28)
``` go
type ABImage struct {
    Digest    string    `json:"digest"`
    Timestamp time.Time `json:"timestamp"`
    Image     string    `json:"image"`
}

```
The ABImage is the representation of an OCI image used by ABRoot, it
contains the digest, the timestamp and the image name. If you need to
investigate the current ABImage on an ABRoot system, you can find it
at /abimage.abr







### <a name="NewABImage">func</a> [NewABImage](/src/target/image.go?s=976:1038#L36)
``` go
func NewABImage(digest string, image string) (*ABImage, error)
```
NewABImage creates a new ABImage instance and returns a pointer to it,
if the digest is empty, it returns an error


### <a name="NewABImageFromRoot">func</a> [NewABImageFromRoot](/src/target/image.go?s=1814:1857#L56)
``` go
func NewABImageFromRoot() (*ABImage, error)
```
NewABImageFromRoot returns the current ABImage by parsing /abimage.abr, if
it fails, it returns an error (e.g. if the file doesn't exist).
Note for distro maintainers: if the /abimage.abr is not present, it could
mean that the user is running an older version of ABRoot (pre v2) or the
root state is corrupted. In the latter case, generating a new ABImage should
fix the issue, Digest and Timestamp can be random, but Image should reflect
an existing image on the configured Docker registry. Anyway, support on this
is not guaranteed, so please don't open issues about this.





### <a name="ABImage.WriteTo">func</a> (\*ABImage) [WriteTo](/src/target/image.go?s=2392:2451#L78)
``` go
func (a *ABImage) WriteTo(dest string, suffix string) error
```
WriteTo writes the json to a destination path, if the suffix is not empty,
it will be appended to the filename




## <a name="ABRollbackResponse">type</a> [ABRollbackResponse](/src/target/system.go?s=2490:2520#L86)
``` go
type ABRollbackResponse string
```
ABRollbackResponse represents the response of a rollback operation










## <a name="ABRootManager">type</a> [ABRootManager](/src/target/root.go?s=893:1090#L28)
``` go
type ABRootManager struct {
    // Partitions is a list of partitions managed by ABRoot
    Partitions []ABRootPartition

    // VarPartition is the partition where /var is mounted
    VarPartition Partition
}

```
ABRootManager exposes methods to manage ABRoot partitions, this includes
getting the present and future partitions, the boot partition, the init
volume (when using LVM Thin-Provisioning), and the other partition. If you
need to operate on an ABRoot partition, you should use this struct, each
partition is a pointer to a Partition struct, which contains methods to
operate on the partition itself







### <a name="NewABRootManager">func</a> [NewABRootManager](/src/target/root.go?s=1492:1530#L49)
``` go
func NewABRootManager() *ABRootManager
```
NewABRootManager creates a new ABRootManager





### <a name="ABRootManager.GetBoot">func</a> (\*ABRootManager) [GetBoot](/src/target/root.go?s=6234:6300#L202)
``` go
func (a *ABRootManager) GetBoot() (partition Partition, err error)
```
GetBoot gets the boot partition from the current device




### <a name="ABRootManager.GetFuture">func</a> (\*ABRootManager) [GetFuture](/src/target/root.go?s=4568:4642#L151)
``` go
func (a *ABRootManager) GetFuture() (partition ABRootPartition, err error)
```
GetFuture gets the future partition




### <a name="ABRootManager.GetInit">func</a> (\*ABRootManager) [GetInit](/src/target/root.go?s=6766:6832#L219)
``` go
func (a *ABRootManager) GetInit() (partition Partition, err error)
```
GetInit gets the init volume when using LVM Thin-Provisioning




### <a name="ABRootManager.GetOther">func</a> (\*ABRootManager) [GetOther](/src/target/root.go?s=5077:5150#L167)
``` go
func (a *ABRootManager) GetOther() (partition ABRootPartition, err error)
```
GetOther gets the other partition




### <a name="ABRootManager.GetPartition">func</a> (\*ABRootManager) [GetPartition](/src/target/root.go?s=5703:5792#L186)
``` go
func (a *ABRootManager) GetPartition(label string) (partition ABRootPartition, err error)
```
GetPartition gets a partition by label




### <a name="ABRootManager.GetPartitions">func</a> (\*ABRootManager) [GetPartitions](/src/target/root.go?s=1708:1753#L59)
``` go
func (a *ABRootManager) GetPartitions() error
```
GetPartitions gets the root partitions from the current device




### <a name="ABRootManager.GetPresent">func</a> (\*ABRootManager) [GetPresent](/src/target/root.go?s=4050:4125#L135)
``` go
func (a *ABRootManager) GetPresent() (partition ABRootPartition, err error)
```
GetPresent gets the present partition




### <a name="ABRootManager.IdentifyPartition">func</a> (\*ABRootManager) [IdentifyPartition](/src/target/root.go?s=3365:3460#L116)
``` go
func (a *ABRootManager) IdentifyPartition(partition Partition) (identifiedAs string, err error)
```
IdentifyPartition identifies a partition




### <a name="ABRootManager.IsCurrent">func</a> (\*ABRootManager) [IsCurrent](/src/target/root.go?s=2987:3046#L103)
``` go
func (a *ABRootManager) IsCurrent(partition Partition) bool
```
IsCurrent checks if a partition is the current one




## <a name="ABRootPartition">type</a> [ABRootPartition](/src/target/root.go?s=1152:1442#L37)
``` go
type ABRootPartition struct {
    Label        string // Matches `partLabelA` and `partLabelB` settings entries
    IdentifiedAs string // Either `present` or `future`
    Partition    Partition
    MountPoint   string
    MountOptions string
    Uuid         string
    FsType       string
    Current      bool
}

```
ABRootPartition represents a partition managed by ABRoot










## <a name="ABSystem">type</a> [ABSystem](/src/target/system.go?s=696:1316#L31)
``` go
type ABSystem struct {
    // Checks contains an instance of Checks which allows to perform
    // compatibility checks on the system such as filesystem compatibility,
    // connectivity and root check.
    Checks *Checks

    // RootM contains an instance of the ABRootManager which allows to
    // manage the ABRoot partition scheme.
    RootM *ABRootManager

    // Registry contains an instance of the Registry used to retrieve resources
    // from the configured Docker registry.
    Registry *Registry

    // CurImage contains an instance of ABImage which represents the current
    // image used by the system (abimage.abr).
    CurImage *ABImage
}

```
An ABSystem allows to perform system operations such as upgrades,
package changes and rollback on an ABRoot-compliant system.







### <a name="NewABSystem">func</a> [NewABSystem](/src/target/system.go?s=3276:3313#L106)
``` go
func NewABSystem() (*ABSystem, error)
```
NewABSystem initializes a new ABSystem, which contains all the functions
to perform system operations such as upgrades, package changes and rollback.
It returns a pointer to the initialized ABSystem and an error, if any.





### <a name="ABSystem.AddToCleanUpQueue">func</a> (\*ABSystem) [AddToCleanUpQueue](/src/target/system.go?s=9426:9512#L323)
``` go
func (s *ABSystem) AddToCleanUpQueue(name string, priority int, values ...interface{})
```
AddToCleanUpQueue adds a function to the queue




### <a name="ABSystem.CheckAll">func</a> (\*ABSystem) [CheckAll](/src/target/system.go?s=3685:3720#L128)
``` go
func (s *ABSystem) CheckAll() error
```
CheckAll performs all checks from the Checks struct




### <a name="ABSystem.CheckUpdate">func</a> (\*ABSystem) [CheckUpdate](/src/target/system.go?s=4023:4070#L142)
``` go
func (s *ABSystem) CheckUpdate() (string, bool)
```
CheckUpdate checks if there is an update available




### <a name="ABSystem.CreateStageFile">func</a> (\*ABSystem) [CreateStageFile](/src/target/system.go?s=34555:34597#L1191)
``` go
func (s *ABSystem) CreateStageFile() error
```
CreateStageFile creates the stage file, which is used to determine if
the upgrade can be interrupted or not. If the stage file is present, it
means that the upgrade is in a state where it is still possible to
interrupt it, otherwise it is not. This is useful for third-party
applications like update managers.




### <a name="ABSystem.GenerateCrypttab">func</a> (\*ABSystem) [GenerateCrypttab](/src/target/system.go?s=11100:11158#L384)
``` go
func (s *ABSystem) GenerateCrypttab(rootPath string) error
```
GenerateCrypttab identifies which devices are encrypted and generates
the /etc/crypttab file for the specified root




### <a name="ABSystem.GenerateFstab">func</a> (\*ABSystem) [GenerateFstab](/src/target/system.go?s=10020:10097#L347)
``` go
func (s *ABSystem) GenerateFstab(rootPath string, root ABRootPartition) error
```
GenerateFstab generates a fstab file for the future root




### <a name="ABSystem.GenerateSystemdUnits">func</a> (\*ABSystem) [GenerateSystemdUnits](/src/target/system.go?s=12549:12633#L434)
``` go
func (s *ABSystem) GenerateSystemdUnits(rootPath string, root ABRootPartition) error
```
GenerateSystemdUnits generates systemd units that mount the mutable parts
of the system to their respective mountpoints




### <a name="ABSystem.LockUpgrade">func</a> (\*ABSystem) [LockUpgrade](/src/target/system.go?s=33690:33728#L1163)
``` go
func (s *ABSystem) LockUpgrade() error
```
LockUpgrade creates a lock file, preventing upgrades from proceeding




### <a name="ABSystem.RemoveFromCleanUpQueue">func</a> (\*ABSystem) [RemoveFromCleanUpQueue](/src/target/system.go?s=9681:9735#L332)
``` go
func (s *ABSystem) RemoveFromCleanUpQueue(name string)
```
RemoveFromCleanUpQueue removes a function from the queue




### <a name="ABSystem.RemoveStageFile">func</a> (\*ABSystem) [RemoveStageFile](/src/target/system.go?s=34903:34945#L1204)
``` go
func (s *ABSystem) RemoveStageFile() error
```
RemoveStageFile removes the stage file disabling the ability to interrupt
the upgrade process




### <a name="ABSystem.ResetQueue">func</a> (\*ABSystem) [ResetQueue](/src/target/system.go?s=9895:9926#L342)
``` go
func (s *ABSystem) ResetQueue()
```
ResetQueue resets the queue




### <a name="ABSystem.Rollback">func</a> (\*ABSystem) [Rollback](/src/target/system.go?s=30352:30422#L1045)
``` go
func (s *ABSystem) Rollback() (response ABRollbackResponse, err error)
```
Rollback swaps the master grub files if the current root is not the default




### <a name="ABSystem.RunCleanUpQueue">func</a> (\*ABSystem) [RunCleanUpQueue](/src/target/system.go?s=6647:6702#L234)
``` go
func (s *ABSystem) RunCleanUpQueue(fnName string) error
```
RunCleanUpQueue runs the functions in the queue or only the specified one




### <a name="ABSystem.RunOperation">func</a> (\*ABSystem) [RunOperation](/src/target/system.go?s=15301:15367#L517)
``` go
func (s *ABSystem) RunOperation(operation ABSystemOperation) error
```
RunOperation executes a root-switching operation from the options below:


	UPGRADE:
		Upgrades to a new image, if available,
	FORCE_UPGRADE:
		Forces the upgrade operation, even if no new image is available,
	APPLY:
		Applies package changes, but doesn't update the system.
	INITRAMFS:
		Updates the initramfs for the future root, but doesn't update the system.



### <a name="ABSystem.UnlockUpgrade">func</a> (\*ABSystem) [UnlockUpgrade](/src/target/system.go?s=33993:34033#L1175)
``` go
func (s *ABSystem) UnlockUpgrade() error
```
UnlockUpgrade removes the lock file, allowing upgrades to proceed




### <a name="ABSystem.UpgradeLockExists">func</a> (\*ABSystem) [UpgradeLockExists](/src/target/system.go?s=33415:33458#L1153)
``` go
func (s *ABSystem) UpgradeLockExists() bool
```
UpgradeLockExists checks if the lock file exists and returns a boolean




### <a name="ABSystem.UserLockRequested">func</a> (\*ABSystem) [UserLockRequested](/src/target/system.go?s=33134:33177#L1143)
``` go
func (s *ABSystem) UserLockRequested() bool
```
UserLockRequested checks if the user lock file exists and returns a boolean
note that if the user lock file exists, it means that the user explicitly
requested the upgrade to be locked (using an update manager for example)




## <a name="ABSystemOperation">type</a> [ABSystemOperation](/src/target/system.go?s=2389:2418#L83)
``` go
type ABSystemOperation string
```
ABSystemOperation represents a system operation to be performed by the
ABSystem, must be given as a parameter to the RunOperation function.










## <a name="Checks">type</a> [Checks](/src/target/checks.go?s=623:643#L28)
``` go
type Checks struct{}

```
Represents a Checks struct which contains all the checks which can
be performed one by one or all at once using PerformAllChecks()







### <a name="NewChecks">func</a> [NewChecks](/src/target/checks.go?s=686:710#L31)
``` go
func NewChecks() *Checks
```
NewChecks returns a new Checks struct





### <a name="Checks.CheckCompatibilityFS">func</a> (\*Checks) [CheckCompatibilityFS](/src/target/checks.go?s=1272:1317#L59)
``` go
func (c *Checks) CheckCompatibilityFS() error
```
CheckCompatibilityFS checks if the filesystem is compatible with ABRoot v2
if not, it returns an error. Note that currently only ext4, btrfs and xfs
are supported/tested. Here we assume some utilities are installed, such as
findmnt and lsblk




### <a name="Checks.CheckConnectivity">func</a> (\*Checks) [CheckConnectivity](/src/target/checks.go?s=2387:2429#L98)
``` go
func (c *Checks) CheckConnectivity() error
```
CheckConnectivity checks if the system is connected to the internet




### <a name="Checks.CheckRoot">func</a> (\*Checks) [CheckRoot](/src/target/checks.go?s=2755:2789#L112)
``` go
func (c *Checks) CheckRoot() error
```
CheckRoot checks if the user is root and returns an error if not




### <a name="Checks.PerformAllChecks">func</a> (\*Checks) [PerformAllChecks](/src/target/checks.go?s=774:815#L36)
``` go
func (c *Checks) PerformAllChecks() error
```
PerformAllChecks performs all checks




## <a name="Children">type</a> [Children](/src/target/disk-manager.go?s=1858:2212#L63)
``` go
type Children struct {
    MountPoint   string     `json:"mountpoint"`
    FsType       string     `json:"fstype"`
    Label        string     `json:"label"`
    Uuid         string     `json:"uuid"`
    LogicalName  string     `json:"name"`
    Size         string     `json:"size"`
    MountOptions string     `json:"mountopts"`
    Children     []Children `json:"children"`
}

```
The children a block device or partition may have










## <a name="Chroot">type</a> [Chroot](/src/target/chroot.go?s=587:666#L26)
``` go
type Chroot struct {
    // contains filtered or unexported fields
}

```
Chroot represents a chroot instance, which can be used to run commands
inside a chroot environment







### <a name="NewChroot">func</a> [NewChroot](/src/target/chroot.go?s=1073:1153#L45)
``` go
func NewChroot(root string, rootUuid string, rootDevice string) (*Chroot, error)
```
NewChroot creates a new chroot environment from the given root path and
returns its Chroot instance or an error if something went wrong





### <a name="Chroot.Close">func</a> (\*Chroot) [Close](/src/target/chroot.go?s=2096:2126#L83)
``` go
func (c *Chroot) Close() error
```
Close unmounts all the bind mounts and closes the chroot environment




### <a name="Chroot.Execute">func</a> (\*Chroot) [Execute](/src/target/chroot.go?s=2915:2972#L116)
``` go
func (c *Chroot) Execute(cmd string, args []string) error
```
Execute runs a command in the chroot environment, the command is
a string and the arguments are a list of strings. If an error occurs
it is returned.




### <a name="Chroot.ExecuteCmds">func</a> (\*Chroot) [ExecuteCmds](/src/target/chroot.go?s=3527:3576#L137)
``` go
func (c *Chroot) ExecuteCmds(cmds []string) error
```
ExecuteCmds runs a list of commands in the chroot environment,
stops at the first error




## <a name="DiskManager">type</a> [DiskManager](/src/target/disk-manager.go?s=630:655#L28)
``` go
type DiskManager struct{}

```
DiskManager exposes functions to interact with the system's disks
and partitions (e.g. mount, unmount, get partitions, etc.)







### <a name="NewDiskManager">func</a> [NewDiskManager](/src/target/disk-manager.go?s=2362:2396#L76)
``` go
func NewDiskManager() *DiskManager
```
NewDiskManager creates and returns a pointer to a new DiskManager instance
from which you can interact with the system's disks and partitions





### <a name="DiskManager.GetPartitionByLabel">func</a> (\*DiskManager) [GetPartitionByLabel](/src/target/disk-manager.go?s=2563:2637#L82)
``` go
func (d *DiskManager) GetPartitionByLabel(label string) (Partition, error)
```
GetPartitionByLabel finds a partition by searching for its label.
If no partition can be found with the given label, returns error.




## <a name="GPUInfo">type</a> [GPUInfo](/src/target/specs.go?s=596:659#L31)
``` go
type GPUInfo struct {
    Address     string
    Description string
}

```









## <a name="Grub">type</a> [Grub](/src/target/grub.go?s=712:772#L29)
``` go
type Grub struct {
    PresentRoot string
    FutureRoot  string
}

```
Grub represents a grub instance, it exposes methods to generate a new grub
config compatible with ABRoot, and to check if the system is booted into
the present root or the future root







### <a name="NewGrub">func</a> [NewGrub](/src/target/grub.go?s=3329:3376#L131)
``` go
func NewGrub(bootPart Partition) (*Grub, error)
```
NewGrub creates a new Grub instance





### <a name="Grub.IsBootedIntoPresentRoot">func</a> (\*Grub) [IsBootedIntoPresentRoot](/src/target/grub.go?s=4403:4457#L174)
``` go
func (g *Grub) IsBootedIntoPresentRoot() (bool, error)
```



## <a name="ImageRecipe">type</a> [ImageRecipe](/src/target/image-recipe.go?s=507:620#L22)
``` go
type ImageRecipe struct {
    From    string
    Labels  map[string]string
    Args    map[string]string
    Content string
}

```
An ImageRecipe represents a Dockerfile/Containerfile-like recipe







### <a name="NewImageRecipe">func</a> [NewImageRecipe](/src/target/image-recipe.go?s=703:815#L30)
``` go
func NewImageRecipe(image string, labels map[string]string, args map[string]string, content string) *ImageRecipe
```
NewImageRecipe creates a new ImageRecipe instance and returns a pointer to it





### <a name="ImageRecipe.Write">func</a> (\*ImageRecipe) [Write](/src/target/image-recipe.go?s=1046:1092#L42)
``` go
func (c *ImageRecipe) Write(path string) error
```
Write writes a ImageRecipe to the given path, returning an error if any




## <a name="IntegrityCheck">type</a> [IntegrityCheck](/src/target/integrity.go?s=498:644#L24)
``` go
type IntegrityCheck struct {
    // contains filtered or unexported fields
}

```






### <a name="NewIntegrityCheck">func</a> [NewIntegrityCheck](/src/target/integrity.go?s=802:884#L34)
``` go
func NewIntegrityCheck(root ABRootPartition, repair bool) (*IntegrityCheck, error)
```
NewIntegrityCheck creates a new IntegrityCheck instance for the given root
partition, and returns a pointer to it or an error if something went wrong





### <a name="IntegrityCheck.Repair">func</a> (\*IntegrityCheck) [Repair](/src/target/integrity.go?s=3675:3715#L154)
``` go
func (ic *IntegrityCheck) Repair() error
```
Repair repairs the system




## <a name="Manifest">type</a> [Manifest](/src/target/registry.go?s=865:942#L33)
``` go
type Manifest struct {
    Manifest []byte
    Digest   string
    Layers   []string
}

```
Manifest is the struct used to parse the manifest response from the registry
it contains the manifest itself, the digest and the list of layers. This
should be compatible with most registries, but it's not guaranteed










## <a name="PCSpecs">type</a> [PCSpecs](/src/target/specs.go?s=524:594#L25)
``` go
type PCSpecs struct {
    CPU    string
    GPU    []string
    Memory string
}

```






### <a name="GetPCSpecs">func</a> [GetPCSpecs](/src/target/specs.go?s=1838:1863#L93)
``` go
func GetPCSpecs() PCSpecs
```




## <a name="PackageManager">type</a> [PackageManager](/src/target/packages.go?s=590:650#L31)
``` go
type PackageManager struct {
    // contains filtered or unexported fields
}

```
PackageManager struct







### <a name="NewPackageManager">func</a> [NewPackageManager](/src/target/packages.go?s=1250:1301#L60)
``` go
func NewPackageManager(dryRun bool) *PackageManager
```
NewPackageManager returns a new PackageManager struct





### <a name="PackageManager.Add">func</a> (\*PackageManager) [Add](/src/target/packages.go?s=2509:2555#L117)
``` go
func (p *PackageManager) Add(pkg string) error
```
Add adds a package to the packages.add file




### <a name="PackageManager.ClearUnstagedPackages">func</a> (\*PackageManager) [ClearUnstagedPackages](/src/target/packages.go?s=6359:6413#L252)
``` go
func (p *PackageManager) ClearUnstagedPackages() error
```
ClearUnstagedPackages removes all packages from the unstaged list




### <a name="PackageManager.ExistsInRepo">func</a> (\*PackageManager) [ExistsInRepo](/src/target/packages.go?s=13366:13421#L490)
``` go
func (p *PackageManager) ExistsInRepo(pkg string) error
```



### <a name="PackageManager.GetAddPackages">func</a> (\*PackageManager) [GetAddPackages](/src/target/packages.go?s=4701:4760#L200)
``` go
func (p *PackageManager) GetAddPackages() ([]string, error)
```
GetAddPackages returns the packages in the packages.add file




### <a name="PackageManager.GetAddPackagesString">func</a> (\*PackageManager) [GetAddPackagesString](/src/target/packages.go?s=6626:6699#L258)
``` go
func (p *PackageManager) GetAddPackagesString(sep string) (string, error)
```
GetAddPackagesString returns the packages in the packages.add file as a string




### <a name="PackageManager.GetFinalCmd">func</a> (\*PackageManager) [GetFinalCmd](/src/target/packages.go?s=11410:11482#L432)
``` go
func (p *PackageManager) GetFinalCmd(operation ABSystemOperation) string
```



### <a name="PackageManager.GetRemovePackages">func</a> (\*PackageManager) [GetRemovePackages](/src/target/packages.go?s=4940:5002#L206)
``` go
func (p *PackageManager) GetRemovePackages() ([]string, error)
```
GetRemovePackages returns the packages in the packages.remove file




### <a name="PackageManager.GetRemovePackagesString">func</a> (\*PackageManager) [GetRemovePackagesString](/src/target/packages.go?s=7102:7178#L271)
``` go
func (p *PackageManager) GetRemovePackagesString(sep string) (string, error)
```
GetRemovePackagesString returns the packages in the packages.remove file as a string




### <a name="PackageManager.GetUnstagedPackages">func</a> (\*PackageManager) [GetUnstagedPackages](/src/target/packages.go?s=5196:5269#L212)
``` go
func (p *PackageManager) GetUnstagedPackages() ([]UnstagedPackage, error)
```
GetUnstagedPackages returns the package changes that are yet to be applied




### <a name="PackageManager.GetUnstagedPackagesPlain">func</a> (\*PackageManager) [GetUnstagedPackagesPlain](/src/target/packages.go?s=5860:5929#L235)
``` go
func (p *PackageManager) GetUnstagedPackagesPlain() ([]string, error)
```
GetUnstagedPackagesPlain returns the package changes that are yet to be applied
as strings




### <a name="PackageManager.Remove">func</a> (\*PackageManager) [Remove](/src/target/packages.go?s=3600:3649#L163)
``` go
func (p *PackageManager) Remove(pkg string) error
```
Remove removes a package from the packages.add file




## <a name="Partition">type</a> [Partition](/src/target/disk-manager.go?s=764:1803#L32)
``` go
type Partition struct {
    Label        string
    MountPoint   string
    MountOptions string
    Uuid         string
    FsType       string

    // If standard partition, Device will be the partition's name (e.g. sda1, nvme0n1p1).
    // If LUKS-encrypted or LVM volume, Device will be the name in device-mapper.
    Device string

    // If the partition is LUKS-encrypted or an LVM volume, the logical volume
    // opened in /dev/mapper will be a child of the physical partition in /dev.
    // Otherwise, the partition will be a direct child of the block device, and
    // Parent will be nil.
    //
    // The same logic applies for encrypted LVM volumes. When this is the case,
    // the filesystem hirearchy is as follows:
    //
    //         NAME               FSTYPE
    //   -- sda1                LVM2_member
    //    |-- myVG-myLV         crypto_LUKS
    //      |-- luks-volume     btrfs
    //
    // In this case, the parent of "luks-volume" is "myVG-myLV", which,
    // in turn, has "sda1" as parent. Since "sda1" is a physical partition,
    // its parent is nil.
    Parent *Partition
}

```
Partition represents either a standard partition or a device-mapper
partition, such as an LVM volume










### <a name="Partition.IsDevMapper">func</a> (\*Partition) [IsDevMapper](/src/target/disk-manager.go?s=6407:6445#L217)
``` go
func (p *Partition) IsDevMapper() bool
```
Returns whether the partition is a device-mapper virtual partition




### <a name="Partition.IsEncrypted">func</a> (\*Partition) [IsEncrypted](/src/target/disk-manager.go?s=6533:6571#L222)
``` go
func (p *Partition) IsEncrypted() bool
```
IsEncrypted returns whether the partition is encrypted




### <a name="Partition.Mount">func</a> (\*Partition) [Mount](/src/target/disk-manager.go?s=5173:5224#L168)
``` go
func (p *Partition) Mount(destination string) error
```
Mount mounts a partition to a directory, returning an error if any occurs




### <a name="Partition.Unmount">func</a> (\*Partition) [Unmount](/src/target/disk-manager.go?s=5870:5905#L196)
``` go
func (p *Partition) Unmount() error
```
Unmount unmounts a partition




## <a name="QueuedFunction">type</a> [QueuedFunction](/src/target/system.go?s=1395:1742#L51)
``` go
type QueuedFunction struct {
    // The name of the function to be executed, which must match one of the
    // supported functions in the RunCleanUpQueue function.
    Name string

    // The values to be passed to the function.
    Values []interface{}

    // The priority of the function. Functions with lower numbers will be
    // executed first.
    Priority int
}

```
QueuedFunction represents a function to be executed in the clean up queue










## <a name="Registry">type</a> [Registry](/src/target/registry.go?s=601:637#L26)
``` go
type Registry struct {
    API string
}

```
A Registry instance exposes functions to interact with the configured
Docker registry







### <a name="NewRegistry">func</a> [NewRegistry](/src/target/registry.go?s=1062:1090#L41)
``` go
func NewRegistry() *Registry
```
NewRegistry returns a new Registry instance, exposing functions to
interact with the configured Docker registry





### <a name="Registry.GetManifest">func</a> (\*Registry) [GetManifest](/src/target/registry.go?s=3059:3122#L114)
``` go
func (r *Registry) GetManifest(token string) (*Manifest, error)
```
GetManifest returns the manifest of the image, a token is required
to perform the request and is generated using GetToken()




### <a name="Registry.HasUpdate">func</a> (\*Registry) [HasUpdate](/src/target/registry.go?s=1416:1474#L50)
``` go
func (r *Registry) HasUpdate(digest string) (string, bool)
```
HasUpdate checks if the image/tag from the registry has a different digest
it returns the new digest and a boolean indicating if an update is available




## <a name="UnstagedPackage">type</a> [UnstagedPackage](/src/target/packages.go?s=1139:1191#L55)
``` go
type UnstagedPackage struct {
    Name, Status string
}

```
An unstaged package is a package that is waiting to be applied
to the next root.

Every time a `pkg apply` or `upgrade` operation
is executed, all unstaged packages are consumed and added/removed
in the next root.














- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
