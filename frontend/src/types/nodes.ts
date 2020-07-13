export declare namespace NodeTypes {

  export interface Node {
    domain: string, 
    author: string, 
    created: string, 
    aproved: string, 
    description: string,
    id: string
  }

  export interface CurrentNode {
    host: string,
    token: string,
  }
}