package ast

import "parsejoy"

type Walker struct {

}

type ASTNode struct {
    Attributes Map[string]interface{}
}

func NewASTNode() *ASTNode {
    node *ASTNode = new(ASTNode)
    node.Initialize()
    return node
}

func (self *ASTNode) Initialize(){
    self.Attributes = make(map[string]interface{})
}

func (self *ASTNode) Set(key string,value interface{}) {
    self.Attributes[key] = value
}

func (self *ASTNode) Get(key string) (interface{},bool) {
    return self.Attributes[key]
}

func (self *Walker) Visit(token *parsejoy.L2Token) *ASTNode {
    switch token.Type {
        case 'funcdef': return self.VisitFuncdef()
        case 'classdef': return self.VisitClassdef()
        case 'expr_stmt': return self.VisitExprStmt()
        case 'file_input' : return self.VisitFileInput()
    }
    return nil
}

func (self *Walker) VisitFuncdef(token *parsejoy.L2Token) {
    /*
    Function Definition
    */
}

func (self *Walker) GetChild(token *parsejoy.L2Token, type string) *parsejoy.L2Token {
    return nil
}


func (self *Walker) VisitClassdef(token *parsejoy.L2Token) (*ASTNode, *L2Token) {
    /*
    * Class Definition
    */
    node := NewASTNode()
    node.Set('node_type','funcdef')
    nameNode := self.GetChild(token,'name')
    node.Set('name',nameNode)
    argumentsNode := self.GetChild(token, 'base_classes')
    node.Set('bases',argumentsNode)
    bodyNode := self.GetChild(token,'suite')
    node.Set('body',self.Visit(bodyNode))
}

func (self *Walker) visitExprStmt(token *parsejoy.L2Token) (*ASTNode, *L2Token) {
    /*
    * Expression
    * Assignment
    * Augmented Assignment
    */
    lhs := 
}

func (self *Walker) visitFileInput(token *parsejoy.L2Token) (*ASTNode, *L2Token) {
    /*
    * Module
    */
}

/*
Core operations:

* Find nodes in the tree by search
* Parse

*/