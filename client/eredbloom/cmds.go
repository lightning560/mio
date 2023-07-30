package eredbloom

import (
	"miopkg/log"
)

func (r *Redbloom) BfAdd(key string, item string) (exists bool, err error) {
	exists, err = r.Client.Add(key, item)
	if err != nil {
		log.Errorf("eredbloom add error", log.FieldErr(err), log.FieldKey(key), log.FieldValue(item))
	}
	return
}

func (r *Redbloom) BfExists(key string, item string) (exists bool, err error) {
	exists, err = r.Client.Exists(key, item)
	if err != nil {
		log.Errorf("eredbloom exists error", log.FieldErr(err), log.FieldKey(key), log.FieldValue(item))
	}
	return
}

// BfAddMulti -Adds one or more items to the Bloom Filter, creating the filter if it does not yet exist. args: key - the name of the filter item - One or more items to add
func (r *Redbloom) BfAddMulti(key string, items []string) (res []int64, err error) {
	res, err = r.Client.BfAddMulti(key, items)
	if err != nil {
		log.Errorf("eredbloom BfAddMulti error", log.FieldErr(err), log.FieldKey(key))
	}
	return
}

// BfExistsMulti - Determines if one or more items may exist in the filter or not. args: key - the name of the filter item - one or more items to check
func (r *Redbloom) BfExistsMulti(key string, items []string) (res []int64, err error) {
	res, err = r.Client.BfExistsMulti(key, items)
	if err != nil {
		log.Errorf("eredbloom BfExistsMulti error", log.FieldErr(err), log.FieldKey(key))
	}
	return
}

// BfInsert - This command will add one or more items to the bloom filter, by default creating it if it does not yet exist.
func (r *Redbloom) BfInsert(key string, cap int64, errorRatio float64, expansion int64, noCreate bool, nonScaling bool, items []string) (res []int64, err error) {
	res, err = r.Client.BfInsert(key, cap, errorRatio, expansion, noCreate, nonScaling, items)
	if err != nil {
		log.Errorf("eredbloom insert error", log.FieldErr(err), log.FieldKey(key))
	}
	return
}

func (r *Redbloom) CfAdd(key string, item string) (exists bool, err error) {
	exists, err = r.Client.CfAdd(key, item)
	if err != nil {
		log.Errorf("eredbloom add error", log.FieldErr(err), log.FieldKey(key), log.FieldValue(item))
	}
	return
}

func (r *Redbloom) CfAddMulti(key string, items []string) (res []int64, err error) {
	for _, item := range items {
		exists, err := r.Client.CfAdd(key, item)
		if err != nil {
			log.Warnf("eredbloom add error", log.FieldErr(err), log.FieldKey(key))
		}
		if exists {
			res = append(res, 1)
		} else {
			res = append(res, 0)
		}
	}
	return
}

func (r *Redbloom) CfExists(key string, item string) (exists bool, err error) {
	exists, err = r.Client.CfExists(key, item)
	if err != nil {
		log.Errorf("eredbloom CfExists error", log.FieldErr(err), log.FieldKey(key), log.FieldValue(item))
	}
	return
}

func (r *Redbloom) CfExistsMulti(key string, items []string) (res []int64, err error) {
	for _, item := range items {
		exists, err := r.Client.CfExists(key, item)
		if err != nil {
			log.Errorf("eredbloom CfExistsMulti error", log.FieldErr(err), log.FieldKey(key))
		}
		if exists {
			res = append(res, 1)
		} else {
			res = append(res, 0)
		}
	}
	return
}

// Adds one or more items to a cuckoo filter, allowing the filter to be created with a custom capacity if it does not yet exist.
func (r *Redbloom) CfInsert(key string, cap int64, noCreate bool, items []string) ([]int64, error) {
	res, err := r.Client.CfInsert(key, cap, noCreate, items)
	if err != nil {
		log.Errorf("eredbloom insert error", log.FieldErr(err), log.FieldKey(key))
	}
	return res, err
}
