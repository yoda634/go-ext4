package ext4

import (
	"fmt"
	"io"
	"strconv"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

const (
	BlockGroupDescriptorSize = 64
)

const (
	BgdFlagInodeTableAndBitmapNotInitialized = uint16(0x1)
	BgdFlagBitmapNotInitialized              = uint16(0x2)
	BgdFlagInodeTableZeroed                  = uint16(0x4)
)

type BlockGroupDescriptor struct {
	BgBlockBitmapLo     uint32    /* Blocks bitmap block */
	BgInodeBitmapLo     uint32    /* Inodes bitmap block */
	BgInodeTableLo      uint32    /* Inodes table block */
	BgFreeBlocksCountLo uint16    /* Free blocks count */
	BgFreeInodesCountLo uint16    /* Free inodes count */
	BgUsedDirsCountLo   uint16    /* Directories count */
	BgFlags             uint16    /* EXT4_BG_flags (INODE_UNINIT, etc) */
	BgReserved          [2]uint32 /* Likely block/inode bitmap checksum */
	BgItableUnusedLo    uint16    /* Unused inodes count */
	BgChecksum          uint16    /* crc16(sb_uuid+group+desc) */
	BgBlockBitmapHi     uint32    /* Blocks bitmap block MSB */
	BgInodeBitmapHi     uint32    /* Inodes bitmap block MSB */
	BgInodeTableHi      uint32    /* Inodes table block MSB */
	BgFreeBlocksCountHi uint16    /* Free blocks count MSB */
	BgFreeInodesCountHi uint16    /* Free inodes count MSB */
	BgUsedDirsCountHi   uint16    /* Directories count MSB */
	BgItableUnusedHi    uint16    /* Unused inodes count MSB */
	BgReserved2         [3]uint32
}

func (bgd *BlockGroupDescriptor) Dump() {
	fmt.Printf("BgBlockBitmapHi: %d\n", bgd.BgBlockBitmapHi)
	fmt.Printf("BgBlockBitmapLo: %d\n", bgd.BgBlockBitmapLo)
	fmt.Printf("BgChecksum: %04x\n", bgd.BgChecksum)
	fmt.Printf("BgFlags: %s\n", strconv.FormatInt(int64(bgd.BgFlags), 2))
	fmt.Printf("BgFreeBlocksCountHi: %d\n", bgd.BgFreeBlocksCountHi)
	fmt.Printf("BgFreeBlocksCountLo: %d\n", bgd.BgFreeBlocksCountLo)
	fmt.Printf("BgFreeInodesCountHi: %d\n", bgd.BgFreeInodesCountHi)
	fmt.Printf("BgFreeInodesCountLo: %d\n", bgd.BgFreeInodesCountLo)
	fmt.Printf("BgInodeBitmapHi: %d\n", bgd.BgInodeBitmapHi)
	fmt.Printf("BgInodeBitmapLo: %d\n", bgd.BgInodeBitmapLo)
	fmt.Printf("BgInodeTableHi: %d\n", bgd.BgInodeTableHi)
	fmt.Printf("BgInodeTableLo: %d\n", bgd.BgInodeTableLo)
	fmt.Printf("BgItableUnusedHi: %d\n", bgd.BgItableUnusedHi)
	fmt.Printf("BgItableUnusedLo: %d\n", bgd.BgItableUnusedLo)
	fmt.Printf("BgReserved2: %x\n", bgd.BgReserved2)
	fmt.Printf("BgReserved: %x\n", bgd.BgReserved)
	fmt.Printf("BgUsedDirsCountHi: %d\n", bgd.BgUsedDirsCountHi)
	fmt.Printf("BgUsedDirsCountLo: %d\n", bgd.BgUsedDirsCountLo)
}

func (bgd *BlockGroupDescriptor) IsInodeTableAndBitmapNotInitialized() bool {
	return (bgd.BgFlags & BgdFlagInodeTableAndBitmapNotInitialized) > 0
}

func (bgd *BlockGroupDescriptor) IsBitmapNotInitialized() bool {
	return (bgd.BgFlags & BgdFlagBitmapNotInitialized) > 0
}

func (bgd *BlockGroupDescriptor) IsInodeTableZeroed() bool {
	return (bgd.BgFlags & BgdFlagInodeTableZeroed) > 0
}

func ParseBlockGroupDescriptor(r io.Reader) (bgd *BlockGroupDescriptor, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	bgd = new(BlockGroupDescriptor)

	err = binary.Read(r, binary.LittleEndian, bgd)
	log.PanicIf(err)

	return bgd, nil
}