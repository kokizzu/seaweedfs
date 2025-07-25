//go:build ydb
// +build ydb

package ydb

import (
	"fmt"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type FileMeta struct {
	DirHash   int64  `ydb:"type:int64"`
	Name      string `ydb:"type:utf8"`
	Directory string `ydb:"type:utf8"`
	Meta      []byte `ydb:"type:string"`
}

type FileMetas []FileMeta

func (fm *FileMeta) queryParameters(ttlSec int32) *table.QueryParameters {
	var expireAtValue types.Value
	if ttlSec > 0 {
		expireAtValue = types.OptionalValue(types.Uint32Value(uint32(ttlSec)))
	} else {
		expireAtValue = types.NullValue(types.TypeUint32)
	}
	return table.NewQueryParameters(
		table.ValueParam("$dir_hash", types.Int64Value(fm.DirHash)),
		table.ValueParam("$directory", types.UTF8Value(fm.Directory)),
		table.ValueParam("$name", types.UTF8Value(fm.Name)),
		table.ValueParam("$meta", types.BytesValue(fm.Meta)),
		table.ValueParam("$expire_at", expireAtValue))
}

func (store *YdbStore) createTableOptions() []options.CreateTableOption {
	columnUnit := options.TimeToLiveUnitSeconds
	return []options.CreateTableOption{
		options.WithColumn("dir_hash", types.TypeInt64),
		options.WithColumn("directory", types.TypeUTF8),
		options.WithColumn("name", types.TypeUTF8),
		options.WithColumn("meta", types.TypeString),
		options.WithColumn("expire_at", types.Optional(types.TypeUint32)),
		options.WithPrimaryKeyColumn("dir_hash", "directory", "name"),
		options.WithTimeToLiveSettings(options.TimeToLiveSettings{
			ColumnName: "expire_at",
			ColumnUnit: &columnUnit,
			Mode:       options.TimeToLiveModeValueSinceUnixEpoch},
		),
		options.WithPartitioningSettings(
			options.WithPartitioningBy([]string{"dir_hash", "name"}),
			options.WithPartitioningBySize(store.partitionBySizeEnabled),
			options.WithPartitionSizeMb(store.partitionSizeMb),
			options.WithPartitioningByLoad(store.partitionByLoadEnabled),
			options.WithMinPartitionsCount(store.minPartitionsCount),
			options.WithMaxPartitionsCount(store.maxPartitionsCount),
		),
	}
}

func withPragma(prefix *string, query string) *string {
	queryWithPragma := fmt.Sprintf(query, *prefix)
	return &queryWithPragma
}
