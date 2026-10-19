package api

import (
	"hash/fnv"
	"strconv"
)

// photoURL builds the public URL of a stored photo. The version query parameter
// is derived from the photo content, so the URL changes whenever the photo is
// replaced and clients never keep showing a stale cached image.
func photoURL(prefix, id string, photo []byte) *string {
	if len(photo) == 0 {
		return nil
	}
	h := fnv.New32a()
	_, _ = h.Write(photo)
	url := prefix + id + photoPathSuffix + "?v=" + strconv.FormatUint(uint64(h.Sum32()), 16)
	return &url
}
