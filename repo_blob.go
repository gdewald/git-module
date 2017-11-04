package git

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type revisionId string

func (repo *Repository) GetBlobInPath(blobIdStr string, blobPath string) (blob *Blob, err error) {
	blobId, err := NewIDFromString(blobIdStr)
	if err != nil {
		return nil, err
	}

	return repo.getBlobInPath(blobId, blobPath)
}

// Finds blob matching the provided hash sum and path. Blob from the first found tree returned.
func (repo *Repository) getBlobInPath(blobId sha1, blobPath string) (blob *Blob, err error) {
	log("searching revs of blob %v in path %s", blobId, blobPath)

	pipeReader, pipeWriter := io.Pipe()

	done := make(chan struct{})

	scanResult := struct {
		Result *Blob
		Err    error
	}{}

	go func() {
		// Send scanResult after function completion.
		defer func() {
			if err := recover(); err != nil {
				scanResult.Err = err.(error)
			}
			done <- struct{}{}
		}()

		scanner := bufio.NewScanner(pipeReader)
		log("scanning revs for blob in path %s", blobPath)
		for scanner.Scan() {
			revIdStr := strings.TrimSpace(scanner.Text())
			log("getting tree at rev %s", revIdStr)

			var treeAtRev *Tree
			treeAtRev, err = repo.GetTree(revIdStr)

			if err != nil {
				log("failed to get tree at rev %s", revIdStr)
				break
			}

			var treeBlob *Blob
			log("getting blob from tree at path %s", blobPath)
			treeBlob, err = treeAtRev.GetBlobByPath(blobPath)

			if err != nil {
				log("failed to get blob from tree at path %s", blobPath)
				break
			}

			log("blob ID in rev %s for path %s is %v", revIdStr, blobPath, treeBlob.ID)
			if treeBlob.ID.Equal(blobId) {
				log("found matching blob with ID %v", treeBlob.ID)
				scanResult.Result = treeBlob
				break
			}
		}

		if scannerErr := scanner.Err(); scannerErr != nil {
			scanResult.Err = scannerErr
		} else if err != nil {
			scanResult.Err = err
		} else if scanResult.Result == nil {
			scanResult.Err = ErrNotExist{}
		}

		pipeReader.Close()
	}()

	stderr := new(bytes.Buffer)
	log("searching all revisions of %s", blobPath)
	err = NewCommand("rev-list", "--all", "--", ":"+blobPath).RunInDirPipeline(repo.Path, pipeWriter, stderr)
	pipeWriter.Close()

	<-done

	if err != nil && err != io.ErrClosedPipe {
		return blob, concatenateError(err, stderr.String())
	} else {
		return scanResult.Result, scanResult.Err
	}

	return blob, err
}
