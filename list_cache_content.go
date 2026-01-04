package gopdf

import "io"

type listCacheContent struct {
	caches []ICacheContent
}

func (l *listCacheContent) append(cache ICacheContent) {
	l.caches = append(l.caches, cache)
}

func (l *listCacheContent) write(w io.Writer) error {
	for _, cache := range l.caches {
		if err := cache.write(w); err != nil {
			return err
		}
	}
	return nil
}
