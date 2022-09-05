export interface NodeData {
	props: object
	name: string
	environment: string
	ip: string
	runList: string
	attributes: string
}

export interface NodeDataProps {
	nodeData: NodeData
}

export interface NodeDataListProps {
	nodeDataList: NodeData[]
}
