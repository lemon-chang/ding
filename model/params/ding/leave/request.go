package leave

type RequestDingLeave struct {
	UseridList string `json:"userid_list"` // 待查询用户的ID列表，每次最多100个。
	StartTime  int64  `json:"start_time"`  // 开始时间 ，Unix时间戳，支持最多180天的查询。
	EndTime    int64  `json:"end_time"`    // 结束时间，Unix时间戳，支持最多180天的查询。
	Offset     int    `json:"offset"`      // 支持分页查询，与size参数同时设置时才生效，此参数代表偏移量，偏移量从0开始。
	Size       int    `json:"size"`        // 支持分页查询，与offset参数同时设置时才生效，此参数代表分页大小，最大20。
}
