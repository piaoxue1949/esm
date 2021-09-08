/*
Copyright 2016 Medcl (m AT medcl.net)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import "sync"

type Indexes map[string]interface{}

type Document struct {
	Index   string                 `json:"_index,omitempty"`
	Type    string                 `json:"_type,omitempty"`
	Id      string                 `json:"_id,omitempty"`
	source  map[string]interface{} `json:"_source,omitempty"`
	Routing string                 `json:"routing,omitempty"` //after 6, only `routing` was supported
}

type Scroll struct {
	Took     int    `json:"took,omitempty"`
	ScrollId string `json:"_scroll_id,omitempty"`
	TimedOut bool   `json:"timed_out,omitempty"`
	Hits     struct {
		MaxScore float32       `json:"max_score,omitempty"`
		Total    int           `json:"total,omitempty"`
		Docs     []interface{} `json:"hits,omitempty"`
	} `json:"hits"`
	Shards struct {
		Total      int `json:"total,omitempty"`
		Successful int `json:"successful,omitempty"`
		Skipped    int `json:"skipped,omitempty"`
		Failed     int `json:"failed,omitempty"`
		Failures   []struct {
			Shard  int         `json:"shard,omitempty"`
			Index  string      `json:"index,omitempty"`
			Status int         `json:"status,omitempty"`
			Reason interface{} `json:"reason,omitempty"`
		} `json:"failures,omitempty"`
	} `json:"_shards,omitempty"`
}

type ScrollV7 struct {
	Scroll
	Hits struct {
		MaxScore float32 `json:"max_score,omitempty"`
		Total    struct {
			Value    int    `json:"value,omitempty"`
			Relation string `json:"relation,omitempty"`
		} `json:"total,omitempty"`
		Docs []interface{} `json:"hits,omitempty"`
	} `json:"hits"`
}

type ClusterVersion struct {
	Name        string `json:"name,omitempty"`
	ClusterName string `json:"cluster_name,omitempty"`
	Version     struct {
		Number        string `json:"number,omitempty"`
		LuceneVersion string `json:"lucene_version,omitempty"`
	} `json:"version,omitempty"`
}

type ClusterHealth struct {
	Name   string `json:"cluster_name,omitempty"`
	Status string `json:"status,omitempty"`
}

//{"took":23,"errors":true,"items":[{"create":{"_index":"mybank3","_type":"my_doc2","_id":"AWz8rlgUkzP-cujdA_Fv","status":409,"error":{"type":"version_conflict_engine_exception","reason":"[AWz8rlgUkzP-cujdA_Fv]: version conflict, document already exists (current version [1])","index_uuid":"w9JZbJkfSEWBI-uluWorgw","shard":"0","index":"mybank3"}}},{"create":{"_index":"mybank3","_type":"my_doc4","_id":"AWz8rpF2kzP-cujdA_Fx","status":400,"error":{"type":"illegal_argument_exception","reason":"Rejecting mapping update to [mybank3] as the final mapping would have more than 1 type: [my_doc2, my_doc4]"}}},{"create":{"_index":"mybank3","_type":"my_doc1","_id":"AWz8rjpJkzP-cujdA_Fu","status":400,"error":{"type":"illegal_argument_exception","reason":"Rejecting mapping update to [mybank3] as the final mapping would have more than 1 type: [my_doc2, my_doc1]"}}},{"create":{"_index":"mybank3","_type":"my_doc3","_id":"AWz8rnbckzP-cujdA_Fw","status":400,"error":{"type":"illegal_argument_exception","reason":"Rejecting mapping update to [mybank3] as the final mapping would have more than 1 type: [my_doc2, my_doc3]"}}},{"create":{"_index":"mybank3","_type":"my_doc5","_id":"AWz8rrsEkzP-cujdA_Fy","status":400,"error":{"type":"illegal_argument_exception","reason":"Rejecting mapping update to [mybank3] as the final mapping would have more than 1 type: [my_doc2, my_doc5]"}}},{"create":{"_index":"mybank3","_type":"doc","_id":"3","status":400,"error":{"type":"illegal_argument_exception","reason":"Rejecting mapping update to [mybank3] as the final mapping would have more than 1 type: [my_doc2, doc]"}}}]}
type BulkResponse struct {
	Took   int                 `json:"took,omitempty"`
	Errors bool                `json:"errors,omitempty"`
	Items  []map[string]Action `json:"items,omitempty"`
}

type Action struct {
	Index  string      `json:"_index,omitempty"`
	Type   string      `json:"_type,omitempty"`
	Id     string      `json:"_id,omitempty"`
	Status int         `json:"status,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

type Migrator struct {
	FlushLock   sync.Mutex
	DocChan     chan map[string]interface{}
	SourceESAPI ESAPI
	TargetESAPI ESAPI
	SourceAuth  *Auth
	TargetAuth  *Auth
	Config      *Config
}

type Config struct {

	// config options
	SourceEs            string `short:"s" long:"source"  description:"源elasticsearch实例, ie: http://localhost:9200"`
	Query               string `short:"q" long:"query"  description:"根据源elasticsearch实例进行查询，在迁移之前过滤数据, ie: name:medcl"`
	TargetEs            string `short:"d" long:"dest"    description:"目标elasticsearch实例, ie: http://localhost:9201"`
	SourceEsAuthStr     string `short:"m" long:"source_auth"  description:"源elasticsearch实例的基本认证, ie: user:pass"`
	TargetEsAuthStr     string `short:"n" long:"dest_auth"  description:"目标elasticsearch实例的基本验证, ie: user:pass"`
	DocBufferCount      int    `short:"c" long:"count"   description:"在scroll请求中每次文档的数量: ie:10000 " default:"10000"`
	Workers             int    `short:"w" long:"workers" description:"批量工作进程数量" default:"1"`
	BulkSizeInMB        int    `short:"b" long:"bulk_size" description:"bulk size，单位MB" default:"5"`
	ScrollTime          string `short:"t" long:"time"    description:"scroll time" default:"1m"`
	ScrollSliceSize     int    `long:"sliced_scroll_size"    description:"scroll切片的大小，要使它工作，大小应该是> 1" default:"1"`
	RecreateIndex       bool   `short:"f" long:"force"   description:"复制前删除目标索引"`
	CopyAllIndexes      bool   `short:"a" long:"all"     description:"复制所有索引(包括.和_开始的)"`
	CopyIndexSettings   bool   `long:"copy_settings"          description:"从源复制索引settings"`
	CopyIndexMappings   bool   `long:"copy_mappings"          description:"从源复制索引mappings"`
	ShardsCount         int    `long:"shards"            description:"在新创建的索引上设置多个分片"`
	SourceIndexNames    string `short:"x" long:"src_indexes" description:"要复制的索引名称，支持正则表达式和逗号分隔列表" default:"_all"`
	TargetIndexName     string `short:"y" long:"dest_index" description:"要保存的索引名称，只允许一个索引名，如果没有指定，将使用原始的索引名" default:""`
	OverrideTypeName    string `short:"u" long:"type_override" description:"覆盖type名称" default:""`
	WaitForGreen        bool   `long:"green"             description:"等待两台主机的集群状态都变为绿色后，再执行转储操作。否则黄色就可以了"`
	LogLevel            string `short:"v" long:"log"            description:"设置日志级别,选择:trace,debug,info,warn,error"  default:"INFO"`
	DumpOutFile         string `short:"o" long:"output_file"            description:"源索引导出的文件名称" `
	DumpInputFile       string `short:"i" long:"input_file"            description:"要导入到索引的文件名称" `
	InputFileType       string `long:"input_file_type"                 description:"输入文件的数据类型，选项:dump、json_line、json_array、log_line" default:"dump" `
	SourceProxy         string `long:"source_proxy"            description:"源HTTP代理, ie: http://127.0.0.1:8080"`
	TargetProxy         string `long:"dest_proxy"            description:"目标HTTP代理, ie: http://127.0.0.1:8080"`
	Refresh             bool   `long:"refresh"                 description:"迁移完成后刷新"`
	Fields              string `long:"fields"                 description:"源字段筛选，逗号分隔, ie: col1,col2,col3,..." `
	RenameFields        string `long:"rename"                 description:"重命名源字段，逗号分隔, ie: _type:type, name:myname" `
	LogstashEndpoint    string `short:"l"  long:"logstash_endpoint"    description:"目标logstash的TCP地址, ie: 127.0.0.1:5055" `
	LogstashSecEndpoint bool   `long:"secured_logstash_endpoint"    description:"由TLS保护的目标logstash的tcp地址" `
	//TestLevel  			string `long:"test_level"    description:"target logstash tcp endpoint was secured by TLS" `
	//TestEnvironment  string `long:"test_environment"    description:"target logstash tcp endpoint was secured by TLS" `

	RepeatOutputTimes int  `long:"repeat_times"            description:"从源重复数据N次至dest输出，需要使用regenerate_id重新生成id"`
	RegenerateID      bool `short:"r" long:"regenerate_id"   description:"为文档重新生成id，这将覆盖数据源中现有的文档id"`
}

type Auth struct {
	User string
	Pass string
}
