package ext4

import (
	"io"
	"path"

	"github.com/dsoprea/go-logging"
)

type directoryWalkQueueItem struct {
	fullDirectoryPath string
	inode             *Inode
	directoryBrowser  *DirectoryBrowser
}

// DirectoryWalk provides full directory-structure recursion.
type DirectoryWalk struct {
	rs                   io.ReadSeeker
	blockGroupDescriptor *BlockGroupDescriptor
	inodeQueue           []directoryWalkQueueItem
}

func NewDirectoryWalk(rs io.ReadSeeker, bgd *BlockGroupDescriptor, rootInodeNumber int) (dw *DirectoryWalk, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	dw = &DirectoryWalk{
		rs:                   rs,
		blockGroupDescriptor: bgd,
	}

	inode, db, err := dw.openInode(rootInodeNumber)
	log.PanicIf(err)

	dwqi := directoryWalkQueueItem{
		inode:            inode,
		directoryBrowser: db,
	}

	dw.inodeQueue = []directoryWalkQueueItem{dwqi}

	return dw, nil
}

func (dw *DirectoryWalk) openInode(inodeNumber int) (inode *Inode, db *DirectoryBrowser, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	inode, err = NewInodeWithReadSeeker(dw.blockGroupDescriptor, dw.rs, inodeNumber)
	log.PanicIf(err)

	db = NewDirectoryBrowser(dw.rs, inode)

	return inode, db, nil
}

// Next steps through the entire tree starting at the given root inode, one
// entry at a time. We guarantee that all adjacent entries will be processed
// adjacently. This will not return the "." and ".." entries.
func (dw *DirectoryWalk) Next() (fullPath string, de *DirectoryEntry, err error) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	for {
		if len(dw.inodeQueue) == 0 {
			return "", nil, io.EOF
		}

		// Keep popping entries off the current directory until we've read
		// everything.
		dwqi := dw.inodeQueue[0]

		de, err := dwqi.directoryBrowser.Next()
		if err == io.EOF {
			// No more entries.

			dw.inodeQueue = dw.inodeQueue[1:]
			continue
		} else if err != nil {
			log.Panic(err)
		}

		// There was at least one more entry.

		filename := de.Name()

		// Skip the special files. We have a handle on things and they're not
		// especially useful since they don't actually contain the directory
		// names.
		if filename == "." || filename == ".." {
			continue
		}

		var fullFilepath string
		if dwqi.fullDirectoryPath == "" {
			fullFilepath = filename
		} else {
			fullFilepath = path.Join(dwqi.fullDirectoryPath, filename)
		}

		// If it's a directory, enqueue it).
		if de.IsDirectory() && filename != "lost+found" {
			childInode, childDb, err := dw.openInode(int(de.data.Inode))
			log.PanicIf(err)

			newDwqi := directoryWalkQueueItem{
				fullDirectoryPath: fullFilepath,
				inode:             childInode,
				directoryBrowser:  childDb,
			}

			dw.inodeQueue = append(dw.inodeQueue, newDwqi)
		}

		return fullFilepath, de, nil
	}
}