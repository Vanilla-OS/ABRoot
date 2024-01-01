

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
* [func GetRepoContentsForPkg(pkg string) (map[string]any, error)](#GetRepoContentsForPkg)
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
* [func PrintVerbose(msg string, args ...interface{})](#PrintVerbose)
* [func PrintVerboseNoLog(msg string, args ...interface{})](#PrintVerboseNoLog)
* [func RetrieveImageForRoot(root string) (string, error)](#RetrieveImageForRoot)
* [func RootCheck(display bool) bool](#RootCheck)
* [func WriteDiff(destFile string, diffLines []byte) error](#WriteDiff)
* [type ABImage](#ABImage)
  * [func NewABImage(digest string, image string) (*ABImage, error)](#NewABImage)
  * [func NewABImageFromRoot() (*ABImage, error)](#NewABImageFromRoot)
  * [func (a *ABImage) WriteTo(dest string, suffix string) error](#ABImage.WriteTo)
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
  * [func (s *ABSystem) MergeUserEtcFiles(oldUpperEtc, newLowerEtc, newUpperEtc string) error](#ABSystem.MergeUserEtcFiles)
  * [func (s *ABSystem) RemoveFromCleanUpQueue(name string)](#ABSystem.RemoveFromCleanUpQueue)
  * [func (s *ABSystem) RemoveStageFile() error](#ABSystem.RemoveStageFile)
  * [func (s *ABSystem) ResetQueue()](#ABSystem.ResetQueue)
  * [func (s *ABSystem) RunCleanUpQueue(fnName string) error](#ABSystem.RunCleanUpQueue)
  * [func (s *ABSystem) RunOperation(operation ABSystemOperation) error](#ABSystem.RunOperation)
  * [func (s *ABSystem) SyncUpperEtc(newEtc string) error](#ABSystem.SyncUpperEtc)
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
    UPGRADE           = "upgrade"
    FORCE_UPGRADE     = "force-upgrade"
    DRY_RUN_UPGRADE   = "dry-run-upgrade"
    APPLY             = "package-apply"
    DRY_RUN_APPLY     = "dry-run-package-apply"
    INITRAMFS         = "initramfs"
    DRY_RUN_INITRAMFS = "dry-run-initramfs"
)
```
``` go
const (
    MountUnitDir = "/etc/systemd/system"
)
```

## <a name="pkg-variables">Variables</a>
``` go
var (
    ErrNoUpdate error = errors.New("no update available")
)
```
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



## <a name="AtomicSwap">func</a> [AtomicSwap](/src/target/atomic-io.go?s=686:724#L26)
``` go
func AtomicSwap(src, dst string) error
```
atomicSwap allows swapping 2 files or directories in-place and atomically,
using the renameat2 syscall. This should be used instead of os.Rename,
which is not atomic at all.



## <a name="BaseImagePackageDiff">func</a> [BaseImagePackageDiff](/src/target/package-diff.go?s=732:864#L30)
``` go
func BaseImagePackageDiff(currentDigest, newDigest string) (
    added, upgraded, downgraded, removed []diff.PackageDiff,
    err error,
)
```
BaseImagePackageDiff retrieves the added, removed, upgraded and downgraded
base packages (the ones bundled with the image).



## <a name="CopyFile">func</a> [CopyFile](/src/target/utils.go?s=1376:1416#L74)
``` go
func CopyFile(source, dest string) error
```
CopyFile copies a file from source to dest



## <a name="DeleteImageForRoot">func</a> [DeleteImageForRoot](/src/target/oci.go?s=4047:4089#L164)
``` go
func DeleteImageForRoot(root string) error
```
DeleteImageForRoot deletes the image created for the provided root ("vos-a"|"vos-b")



## <a name="DiffFiles">func</a> [DiffFiles](/src/target/diff.go?s=1442:1501#L57)
``` go
func DiffFiles(sourceFile, destFile string) ([]byte, error)
```
DiffFiles returns the diff lines between source and dest files.



## <a name="FindImageWithLabel">func</a> [FindImageWithLabel](/src/target/oci.go?s=2711:2769#L114)
``` go
func FindImageWithLabel(key, value string) (string, error)
```
FindImageWithLabel returns the name of the first image containinig the provided key-value pair
or an empty string if none was found



## <a name="GetLogFile">func</a> [GetLogFile](/src/target/logging.go?s=1980:2006#L98)
``` go
func GetLogFile() *os.File
```


## <a name="GetRepoContentsForPkg">func</a> [GetRepoContentsForPkg](/src/target/packages.go?s=14199:14261#L520)
``` go
func GetRepoContentsForPkg(pkg string) (map[string]any, error)
```
GetRepoContentsForPkg retrieves package information from the repository API



## <a name="GetToken">func</a> [GetToken](/src/target/registry.go?s=1685:1716#L70)
``` go
func GetToken() (string, error)
```
GetToken generates a token using the provided tokenURL and returns it



## <a name="IsVerbose">func</a> [IsVerbose](/src/target/logging.go?s=1338:1359#L64)
``` go
func IsVerbose() bool
```


## <a name="KargsBackup">func</a> [KargsBackup](/src/target/kargs.go?s=1884:1908#L84)
``` go
func KargsBackup() error
```
KargsBackup makes a backup of the current kargs file



## <a name="KargsEdit">func</a> [KargsEdit](/src/target/kargs.go?s=3643:3673#L159)
``` go
func KargsEdit() (bool, error)
```
KargsEdit copies the kargs file to a temporary file and opens it in the
user's preferred editor by querying the $EDITOR environment variable.
Once closed, its contents are written back to the main kargs file.
This function returns a boolean parameter indicating whether any changes
were made to the kargs file.



## <a name="KargsFormat">func</a> [KargsFormat](/src/target/kargs.go?s=2768:2816#L123)
``` go
func KargsFormat(content string) (string, error)
```
KargsFormat formats the contents of the kargs file, ensuring that
there are no duplicate entries, multiple spaces or trailing newline



## <a name="KargsRead">func</a> [KargsRead](/src/target/kargs.go?s=2275:2307#L103)
``` go
func KargsRead() (string, error)
```
KargsRead reads the content of the kargs file



## <a name="KargsWrite">func</a> [KargsWrite](/src/target/kargs.go?s=1243:1280#L53)
``` go
func KargsWrite(content string) error
```
KargsWrite makes a backup of the current kargs file and then
writes the new content to it



## <a name="LogToFile">func</a> [LogToFile](/src/target/logging.go?s=1739:1792#L85)
``` go
func LogToFile(msg string, args ...interface{}) error
```


## <a name="MergeDiff">func</a> [MergeDiff](/src/target/diff.go?s=539:602#L23)
``` go
func MergeDiff(firstFile, secondFile, destination string) error
```
MergeDiff merges the diff lines between the first and second files into destination



## <a name="OciExportRootFs">func</a> [OciExportRootFs](/src/target/oci.go?s=691:796#L29)
``` go
func OciExportRootFs(buildImageName string, imageRecipe *ImageRecipe, transDir string, dest string) error
```
OciExportRootFs generates a rootfs from a image recipe file



## <a name="OverlayPackageDiff">func</a> [OverlayPackageDiff](/src/target/package-diff.go?s=2565:2664#L88)
``` go
func OverlayPackageDiff() (
    added, upgraded, downgraded, removed []diff.PackageDiff,
    err error,
)
```
OverlayPackageDiff retrieves the added, removed, upgraded and downgraded
overlay packages (the ones added manually via `abroot pkg add`).



## <a name="PrintVerbose">func</a> [PrintVerbose](/src/target/logging.go?s=1464:1514#L70)
``` go
func PrintVerbose(msg string, args ...interface{})
```


## <a name="PrintVerboseNoLog">func</a> [PrintVerboseNoLog](/src/target/logging.go?s=1604:1659#L78)
``` go
func PrintVerboseNoLog(msg string, args ...interface{})
```


## <a name="RetrieveImageForRoot">func</a> [RetrieveImageForRoot](/src/target/oci.go?s=3686:3740#L151)
``` go
func RetrieveImageForRoot(root string) (string, error)
```
RetrieveImageForRoot retrieves the image created for the provided root ("vos-a"|"vos-b")



## <a name="RootCheck">func</a> [RootCheck](/src/target/utils.go?s=690:723#L39)
``` go
func RootCheck(display bool) bool
```


## <a name="WriteDiff">func</a> [WriteDiff](/src/target/diff.go?s=2023:2078#L81)
``` go
func WriteDiff(destFile string, diffLines []byte) error
```
WriteDiff applies the diff lines to the destination file.




## <a name="ABImage">type</a> [ABImage](/src/target/image.go?s=499:635#L25)
``` go
type ABImage struct {
    Digest    string    `json:"digest"`
    Timestamp time.Time `json:"timestamp"`
    Image     string    `json:"image"`
}

```
ABImage struct







### <a name="NewABImage">func</a> [NewABImage](/src/target/image.go?s=680:742#L32)
``` go
func NewABImage(digest string, image string) (*ABImage, error)
```
NewABImage returns a new ABImage struct


### <a name="NewABImageFromRoot">func</a> [NewABImageFromRoot](/src/target/image.go?s=987:1030#L45)
``` go
func NewABImageFromRoot() (*ABImage, error)
```
NewABImageFromRoot returns the current ABImage from /abimage.abr





### <a name="ABImage.WriteTo">func</a> (\*ABImage) [WriteTo](/src/target/image.go?s=1501:1560#L66)
``` go
func (a *ABImage) WriteTo(dest string, suffix string) error
```
WriteTo writes the json to a dest path




## <a name="ABRootManager">type</a> [ABRootManager](/src/target/root.go?s=525:610#L23)
``` go
type ABRootManager struct {
    Partitions   []ABRootPartition
    VarPartition Partition
}

```
ABRootManager represents the ABRoot manager







### <a name="NewABRootManager">func</a> [NewABRootManager](/src/target/root.go?s=1002:1040#L41)
``` go
func NewABRootManager() *ABRootManager
```
NewABRootManager creates a new ABRootManager





### <a name="ABRootManager.GetBoot">func</a> (\*ABRootManager) [GetBoot](/src/target/root.go?s=5676:5742#L194)
``` go
func (a *ABRootManager) GetBoot() (partition Partition, err error)
```
GetBoot gets the boot partition from the current device




### <a name="ABRootManager.GetFuture">func</a> (\*ABRootManager) [GetFuture](/src/target/root.go?s=4037:4111#L143)
``` go
func (a *ABRootManager) GetFuture() (partition ABRootPartition, err error)
```
GetFuture gets the future partition




### <a name="ABRootManager.GetInit">func</a> (\*ABRootManager) [GetInit](/src/target/root.go?s=6201:6267#L211)
``` go
func (a *ABRootManager) GetInit() (partition Partition, err error)
```
GetInit gets the init volume when using LVM Thin-Provisioning




### <a name="ABRootManager.GetOther">func</a> (\*ABRootManager) [GetOther](/src/target/root.go?s=4539:4612#L159)
``` go
func (a *ABRootManager) GetOther() (partition ABRootPartition, err error)
```
GetOther gets the other partition




### <a name="ABRootManager.GetPartition">func</a> (\*ABRootManager) [GetPartition](/src/target/root.go?s=5152:5241#L178)
``` go
func (a *ABRootManager) GetPartition(label string) (partition ABRootPartition, err error)
```
GetPartition gets a partition by label




### <a name="ABRootManager.GetPartitions">func</a> (\*ABRootManager) [GetPartitions](/src/target/root.go?s=1212:1257#L51)
``` go
func (a *ABRootManager) GetPartitions() error
```
GetPartitions gets the root partitions from the current device




### <a name="ABRootManager.GetPresent">func</a> (\*ABRootManager) [GetPresent](/src/target/root.go?s=3526:3601#L127)
``` go
func (a *ABRootManager) GetPresent() (partition ABRootPartition, err error)
```
GetPresent gets the present partition




### <a name="ABRootManager.IdentifyPartition">func</a> (\*ABRootManager) [IdentifyPartition](/src/target/root.go?s=2854:2949#L108)
``` go
func (a *ABRootManager) IdentifyPartition(partition Partition) (identifiedAs string, err error)
```
IdentifyPartition identifies a partition




### <a name="ABRootManager.IsCurrent">func</a> (\*ABRootManager) [IsCurrent](/src/target/root.go?s=2494:2553#L95)
``` go
func (a *ABRootManager) IsCurrent(partition Partition) bool
```
IsCurrent checks if a partition is the current one




## <a name="ABRootPartition">type</a> [ABRootPartition](/src/target/root.go?s=662:952#L29)
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
ABRootPartition represents an ABRoot partition










## <a name="ABSystem">type</a> [ABSystem](/src/target/system.go?s=598:704#L30)
``` go
type ABSystem struct {
    Checks   *Checks
    RootM    *ABRootManager
    Registry *Registry
    CurImage *ABImage
}

```
ABSystem represents the system







### <a name="NewABSystem">func</a> [NewABSystem](/src/target/system.go?s=1496:1533#L68)
``` go
func NewABSystem() (*ABSystem, error)
```
NewABSystem creates a new system





### <a name="ABSystem.AddToCleanUpQueue">func</a> (\*ABSystem) [AddToCleanUpQueue](/src/target/system.go?s=7633:7719#L288)
``` go
func (s *ABSystem) AddToCleanUpQueue(name string, priority int, values ...interface{})
```
AddToCleanUpQueue adds a function to the queue




### <a name="ABSystem.CheckAll">func</a> (\*ABSystem) [CheckAll](/src/target/system.go?s=1903:1938#L90)
``` go
func (s *ABSystem) CheckAll() error
```
CheckAll performs all checks from the Checks struct




### <a name="ABSystem.CheckUpdate">func</a> (\*ABSystem) [CheckUpdate](/src/target/system.go?s=2231:2278#L104)
``` go
func (s *ABSystem) CheckUpdate() (string, bool)
```
CheckUpdate checks if there is an update available




### <a name="ABSystem.CreateStageFile">func</a> (\*ABSystem) [CreateStageFile](/src/target/system.go?s=29265:29307#L1039)
``` go
func (s *ABSystem) CreateStageFile() error
```



### <a name="ABSystem.GenerateCrypttab">func</a> (\*ABSystem) [GenerateCrypttab](/src/target/system.go?s=9297:9355#L349)
``` go
func (s *ABSystem) GenerateCrypttab(rootPath string) error
```
GenerateCrypttab identifies which devices are encrypted and generates
the /etc/crypttab file for the specified root




### <a name="ABSystem.GenerateFstab">func</a> (\*ABSystem) [GenerateFstab](/src/target/system.go?s=8227:8304#L312)
``` go
func (s *ABSystem) GenerateFstab(rootPath string, root ABRootPartition) error
```
GenerateFstab generates a fstab file for the future root




### <a name="ABSystem.GenerateSystemdUnits">func</a> (\*ABSystem) [GenerateSystemdUnits](/src/target/system.go?s=10698:10782#L398)
``` go
func (s *ABSystem) GenerateSystemdUnits(rootPath string, root ABRootPartition) error
```
GenerateSystemdUnits generates systemd units that mount the mutable parts of the system




### <a name="ABSystem.LockUpgrade">func</a> (\*ABSystem) [LockUpgrade](/src/target/system.go?s=28810:28848#L1017)
``` go
func (s *ABSystem) LockUpgrade() error
```



### <a name="ABSystem.MergeUserEtcFiles">func</a> (\*ABSystem) [MergeUserEtcFiles](/src/target/system.go?s=2546:2634#L111)
``` go
func (s *ABSystem) MergeUserEtcFiles(oldUpperEtc, newLowerEtc, newUpperEtc string) error
```
MergeUserEtcFiles merges user-related files from the new lower etc (/.system/etc)
with the old upper etc, if present, saving the result in the new upper etc.




### <a name="ABSystem.RemoveFromCleanUpQueue">func</a> (\*ABSystem) [RemoveFromCleanUpQueue](/src/target/system.go?s=7888:7942#L297)
``` go
func (s *ABSystem) RemoveFromCleanUpQueue(name string)
```
RemoveFromCleanUpQueue removes a function from the queue




### <a name="ABSystem.RemoveStageFile">func</a> (\*ABSystem) [RemoveStageFile](/src/target/system.go?s=29505:29547#L1050)
``` go
func (s *ABSystem) RemoveStageFile() error
```



### <a name="ABSystem.ResetQueue">func</a> (\*ABSystem) [ResetQueue](/src/target/system.go?s=8102:8133#L307)
``` go
func (s *ABSystem) ResetQueue()
```
ResetQueue resets the queue




### <a name="ABSystem.RunCleanUpQueue">func</a> (\*ABSystem) [RunCleanUpQueue](/src/target/system.go?s=4882:4937#L199)
``` go
func (s *ABSystem) RunCleanUpQueue(fnName string) error
```
RunCleanUpQueue runs the functions in the queue or only the specified one




### <a name="ABSystem.RunOperation">func</a> (\*ABSystem) [RunOperation](/src/target/system.go?s=13435:13501#L477)
``` go
func (s *ABSystem) RunOperation(operation ABSystemOperation) error
```
RunOperation executes a root-switching operation from the options below:


	UPGRADE: Upgrades to a new image, if available,
	FORCE_UPGRADE: Forces the upgrade operation, even if no new image is available,
	APPLY: Applies package changes, but doesn't update the system.
	INITRAMFS: Updates the initramfs for the future root, but doesn't update the system.




### <a name="ABSystem.SyncUpperEtc">func</a> (\*ABSystem) [SyncUpperEtc](/src/target/system.go?s=3834:3886#L158)
``` go
func (s *ABSystem) SyncUpperEtc(newEtc string) error
```
SyncUpperEtc syncs the mutable etc directories from /var/lib/abroot/etc




### <a name="ABSystem.UnlockUpgrade">func</a> (\*ABSystem) [UnlockUpgrade](/src/target/system.go?s=29036:29076#L1028)
``` go
func (s *ABSystem) UnlockUpgrade() error
```



### <a name="ABSystem.UpgradeLockExists">func</a> (\*ABSystem) [UpgradeLockExists](/src/target/system.go?s=28613:28656#L1008)
``` go
func (s *ABSystem) UpgradeLockExists() bool
```



### <a name="ABSystem.UserLockRequested">func</a> (\*ABSystem) [UserLockRequested](/src/target/system.go?s=28412:28455#L999)
``` go
func (s *ABSystem) UserLockRequested() bool
```



## <a name="ABSystemOperation">type</a> [ABSystemOperation](/src/target/system.go?s=1116:1145#L57)
``` go
type ABSystemOperation string
```









## <a name="Checks">type</a> [Checks](/src/target/checks.go?s=616:636#L27)
``` go
type Checks struct{}

```
Represents a Checks struct which contains all the checks which can
be performed one by one or all at once using PerformAllChecks()







### <a name="NewChecks">func</a> [NewChecks](/src/target/checks.go?s=679:703#L30)
``` go
func NewChecks() *Checks
```
NewChecks returns a new Checks struct





### <a name="Checks.CheckCompatibilityFS">func</a> (\*Checks) [CheckCompatibilityFS](/src/target/checks.go?s=1074:1119#L55)
``` go
func (c *Checks) CheckCompatibilityFS() error
```
CheckCompatibilityFS checks if the filesystem is compatible




### <a name="Checks.CheckConnectivity">func</a> (\*Checks) [CheckConnectivity](/src/target/checks.go?s=2252:2294#L93)
``` go
func (c *Checks) CheckConnectivity() error
```
CheckConnectivity checks if the system is connected to the internet




### <a name="Checks.CheckRoot">func</a> (\*Checks) [CheckRoot](/src/target/checks.go?s=2598:2632#L107)
``` go
func (c *Checks) CheckRoot() error
```
CheckRoot checks if the user is root




### <a name="Checks.PerformAllChecks">func</a> (\*Checks) [PerformAllChecks](/src/target/checks.go?s=767:808#L35)
``` go
func (c *Checks) PerformAllChecks() error
```
PerformAllChecks performs all checks




## <a name="Children">type</a> [Children](/src/target/disk-manager.go?s=1735:2089#L61)
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










## <a name="Chroot">type</a> [Chroot](/src/target/chroot.go?s=542:621#L25)
``` go
type Chroot struct {
    // contains filtered or unexported fields
}

```
Chroot is a struct which represents a chroot environment







### <a name="NewChroot">func</a> [NewChroot](/src/target/chroot.go?s=753:833#L40)
``` go
func NewChroot(root string, rootUuid string, rootDevice string) (*Chroot, error)
```
NewChroot creates a new chroot environment





### <a name="Chroot.Close">func</a> (\*Chroot) [Close](/src/target/chroot.go?s=1759:1789#L78)
``` go
func (c *Chroot) Close() error
```
Close unmounts all the bind mounts




### <a name="Chroot.Execute">func</a> (\*Chroot) [Execute](/src/target/chroot.go?s=2473:2530#L109)
``` go
func (c *Chroot) Execute(cmd string, args []string) error
```
Execute runs a command in the chroot environment




### <a name="Chroot.ExecuteCmds">func</a> (\*Chroot) [ExecuteCmds](/src/target/chroot.go?s=3078:3127#L130)
``` go
func (c *Chroot) ExecuteCmds(cmds []string) error
```
ExecuteCmds runs a list of commands in the chroot environment,
stops at the first error




## <a name="DiskManager">type</a> [DiskManager](/src/target/disk-manager.go?s=532:557#L27)
``` go
type DiskManager struct{}

```
DiskManager represents a disk







### <a name="NewDiskManager">func</a> [NewDiskManager](/src/target/disk-manager.go?s=2135:2169#L73)
``` go
func NewDiskManager() *DiskManager
```
NewDiskManager creates a new DiskManager





### <a name="DiskManager.GetPartitionByLabel">func</a> (\*DiskManager) [GetPartitionByLabel](/src/target/disk-manager.go?s=2339:2413#L80)
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









## <a name="Grub">type</a> [Grub](/src/target/grub.go?s=519:579#L26)
``` go
type Grub struct {
    PresentRoot string
    FutureRoot  string
}

```






### <a name="NewGrub">func</a> [NewGrub](/src/target/grub.go?s=3047:3094#L126)
``` go
func NewGrub(bootPart Partition) (*Grub, error)
```
NewGrub creates a new Grub instance





### <a name="Grub.IsBootedIntoPresentRoot">func</a> (\*Grub) [IsBootedIntoPresentRoot](/src/target/grub.go?s=4101:4155#L168)
``` go
func (g *Grub) IsBootedIntoPresentRoot() (bool, error)
```



## <a name="ImageRecipe">type</a> [ImageRecipe](/src/target/image-recipe.go?s=439:552#L21)
``` go
type ImageRecipe struct {
    From    string
    Labels  map[string]string
    Args    map[string]string
    Content string
}

```






### <a name="NewImageRecipe">func</a> [NewImageRecipe](/src/target/image-recipe.go?s=605:717#L29)
``` go
func NewImageRecipe(image string, labels map[string]string, args map[string]string, content string) *ImageRecipe
```
NewImageRecipe creates a new ImageRecipe struct





### <a name="ImageRecipe.Write">func</a> (\*ImageRecipe) [Write](/src/target/image-recipe.go?s=907:953#L41)
``` go
func (c *ImageRecipe) Write(path string) error
```
Write writes a ImageRecipe to a path




## <a name="IntegrityCheck">type</a> [IntegrityCheck](/src/target/integrity.go?s=498:644#L24)
``` go
type IntegrityCheck struct {
    // contains filtered or unexported fields
}

```






### <a name="NewIntegrityCheck">func</a> [NewIntegrityCheck](/src/target/integrity.go?s=705:787#L33)
``` go
func NewIntegrityCheck(root ABRootPartition, repair bool) (*IntegrityCheck, error)
```
NewIntegrityCheck creates a new IntegrityCheck instance





### <a name="IntegrityCheck.Repair">func</a> (\*IntegrityCheck) [Repair](/src/target/integrity.go?s=3427:3467#L151)
``` go
func (ic *IntegrityCheck) Repair() error
```
Repair repairs the system




## <a name="Manifest">type</a> [Manifest](/src/target/registry.go?s=585:662#L30)
``` go
type Manifest struct {
    Manifest []byte
    Digest   string
    Layers   []string
}

```
Manifest struct










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





### <a name="PackageManager.Add">func</a> (\*PackageManager) [Add](/src/target/packages.go?s=2539:2585#L117)
``` go
func (p *PackageManager) Add(pkg string) error
```
Add adds a package to the packages.add file




### <a name="PackageManager.ClearUnstagedPackages">func</a> (\*PackageManager) [ClearUnstagedPackages](/src/target/packages.go?s=6419:6473#L252)
``` go
func (p *PackageManager) ClearUnstagedPackages() error
```
ClearUnstagedPackages removes all packages from the unstaged list




### <a name="PackageManager.ExistsInRepo">func</a> (\*PackageManager) [ExistsInRepo](/src/target/packages.go?s=13359:13414#L490)
``` go
func (p *PackageManager) ExistsInRepo(pkg string) error
```



### <a name="PackageManager.GetAddPackages">func</a> (\*PackageManager) [GetAddPackages](/src/target/packages.go?s=4767:4826#L200)
``` go
func (p *PackageManager) GetAddPackages() ([]string, error)
```
GetAddPackages returns the packages in the packages.add file




### <a name="PackageManager.GetAddPackagesString">func</a> (\*PackageManager) [GetAddPackagesString](/src/target/packages.go?s=6672:6745#L258)
``` go
func (p *PackageManager) GetAddPackagesString(sep string) (string, error)
```
GetAddPackages returns the packages in the packages.add file as string




### <a name="PackageManager.GetFinalCmd">func</a> (\*PackageManager) [GetFinalCmd](/src/target/packages.go?s=11425:11497#L432)
``` go
func (p *PackageManager) GetFinalCmd(operation ABSystemOperation) string
```



### <a name="PackageManager.GetRemovePackages">func</a> (\*PackageManager) [GetRemovePackages](/src/target/packages.go?s=5000:5062#L206)
``` go
func (p *PackageManager) GetRemovePackages() ([]string, error)
```
GetRemovePackages returns the packages in the packages.remove file




### <a name="PackageManager.GetRemovePackagesString">func</a> (\*PackageManager) [GetRemovePackagesString](/src/target/packages.go?s=7137:7213#L271)
``` go
func (p *PackageManager) GetRemovePackagesString(sep string) (string, error)
```
GetRemovePackages returns the packages in the packages.remove file as string




### <a name="PackageManager.GetUnstagedPackages">func</a> (\*PackageManager) [GetUnstagedPackages](/src/target/packages.go?s=5250:5323#L212)
``` go
func (p *PackageManager) GetUnstagedPackages() ([]UnstagedPackage, error)
```
GetUnstagedPackages returns the package changes that are yet to be applied




### <a name="PackageManager.GetUnstagedPackagesPlain">func</a> (\*PackageManager) [GetUnstagedPackagesPlain](/src/target/packages.go?s=5917:5986#L235)
``` go
func (p *PackageManager) GetUnstagedPackagesPlain() ([]string, error)
```
GetUnstagedPackagesPlain returns the package changes that are yet to be applied
as strings




### <a name="PackageManager.Remove">func</a> (\*PackageManager) [Remove](/src/target/packages.go?s=3654:3703#L163)
``` go
func (p *PackageManager) Remove(pkg string) error
```
Remove removes a package from the packages.add file




## <a name="Partition">type</a> [Partition](/src/target/disk-manager.go?s=641:1680#L30)
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
Partition represents either a standard partition or a device-mapper partition.










### <a name="Partition.IsDevMapper">func</a> (\*Partition) [IsDevMapper](/src/target/disk-manager.go?s=6054:6092#L211)
``` go
func (p *Partition) IsDevMapper() bool
```
Returns whether the partition is a device-mapper virtual partition




### <a name="Partition.IsEncrypted">func</a> (\*Partition) [IsEncrypted](/src/target/disk-manager.go?s=6180:6218#L216)
``` go
func (p *Partition) IsEncrypted() bool
```
IsEncrypted returns whether the partition is encrypted




### <a name="Partition.Mount">func</a> (\*Partition) [Mount](/src/target/disk-manager.go?s=4757:4808#L161)
``` go
func (p *Partition) Mount(destination string) error
```
Mount mounts a partition to a directory




### <a name="Partition.Unmount">func</a> (\*Partition) [Unmount](/src/target/disk-manager.go?s=5459:5494#L189)
``` go
func (p *Partition) Unmount() error
```
Unmount unmounts a partition




## <a name="QueuedFunction">type</a> [QueuedFunction](/src/target/system.go?s=706:791#L37)
``` go
type QueuedFunction struct {
    Name     string
    Values   []interface{}
    Priority int
}

```









## <a name="Registry">type</a> [Registry](/src/target/registry.go?s=528:564#L25)
``` go
type Registry struct {
    API string
}

```
Registry struct







### <a name="NewRegistry">func</a> [NewRegistry](/src/target/registry.go?s=709:737#L37)
``` go
func NewRegistry() *Registry
```
NewRegistry returns a new Registry struct





### <a name="Registry.GetManifest">func</a> (\*Registry) [GetManifest](/src/target/registry.go?s=2507:2570#L107)
``` go
func (r *Registry) GetManifest(token string) (*Manifest, error)
```
GetManifest returns the manifest of the image




### <a name="Registry.HasUpdate">func</a> (\*Registry) [HasUpdate](/src/target/registry.go?s=977:1035#L45)
``` go
func (r *Registry) HasUpdate(digest string) (string, bool)
```
HasUpdate checks if the image/tag from the registry has a different digest




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
