import { NodeData } from '../types/nodedata'

export async function GetNode(id: string): Promise<NodeData> {
	const response = await fetch(
		`http://localhost/api/node/${id}`
	)
	const nodeData: NodeData = (await response.json()) as NodeData
	return nodeData
}

export async function GetNodes(): Promise<NodeData[]> {
	const response = await fetch(
		'https://jsonplaceholder.typicode.com/nodes?_page=1'
	)
	const nodeList: NodeData[] = (await response.json()) as NodeData[]
	return nodeList
}
