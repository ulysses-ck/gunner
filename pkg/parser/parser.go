package parser

import (
	"strings"
)

type NodeType int

const (
	NODE_MAP NodeType = iota
	NODE_LIST
	NODE_STRING
)

type Node struct {
	Type     NodeType
	Value    string
	Children map[string]*Node
	Items    []*Node
}

func NewMapNode() *Node {
	return &Node{
		Type:     NODE_MAP,
		Children: make(map[string]*Node),
	}
}

func NewListNode() *Node {
	return &Node{
		Type:  NODE_LIST,
		Items: []*Node{},
	}
}

func NewStringNode(value string) *Node {
	return &Node{
		Type:  NODE_STRING,
		Value: value,
	}
}

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) parseKeyValue(tokenValue string) (string, string) {
	parts := strings.SplitN(tokenValue, ":", 2)
	key := strings.TrimSpace(parts[0])
	value := ""
	if len(parts) > 1 {
		value = strings.TrimSpace(parts[1])
	}
	return key, value
}

func (p *Parser) Parse() (*Node, error) {
	root := NewMapNode()
	if err := p.parseMap(root); err != nil {
		return nil, err
	}
	return root, nil
}

func (p *Parser) parseMap(node *Node) error {
	for p.current().Type != TOKEN_EOF && p.current().Type != TOKEN_DEDENT {
		if p.current().Type != TOKEN_KEY {
			break
		}

		key, value := p.parseKeyValue(p.current().Value)
		p.advance()

		// Si hay un valor inline, usarlo
		if value != "" {
			node.Children[key] = NewStringNode(value)
			continue
		}

		// Si no hay valor inline, verificar si viene contenido anidado
		if p.current().Type == TOKEN_INDENT {
			p.advance()

			// Determinar si es un mapa o una lista
			if p.current().Type == TOKEN_DASH {
				listNode := NewListNode()
				if err := p.parseList(listNode); err != nil {
					return err
				}
				node.Children[key] = listNode
			} else if p.current().Type == TOKEN_KEY {
				mapNode := NewMapNode()
				if err := p.parseMap(mapNode); err != nil {
					return err
				}
				node.Children[key] = mapNode
			}

			if p.current().Type == TOKEN_DEDENT {
				p.advance()
			}
		}
	}

	return nil
}

func (p *Parser) parseList(node *Node) error {
	for p.current().Type == TOKEN_DASH {
		p.advance()

		if p.current().Type == TOKEN_STRING {
			node.Items = append(node.Items, NewStringNode(p.current().Value))
			p.advance()
		} else if p.current().Type == TOKEN_KEY {
			// Item complejo (mapa dentro de la lista)
			mapNode := NewMapNode()

			for p.current().Type == TOKEN_KEY {
				key, value := p.parseKeyValue(p.current().Value)
				p.advance()

				if value != "" {
					// Valor inline
					mapNode.Children[key] = NewStringNode(value)
				} else if p.current().Type == TOKEN_INDENT {
					p.advance()

					if p.current().Type == TOKEN_DASH {
						listNode := NewListNode()
						if err := p.parseList(listNode); err != nil {
							return err
						}
						mapNode.Children[key] = listNode
					} else if p.current().Type == TOKEN_KEY {
						nestedMap := NewMapNode()
						if err := p.parseMap(nestedMap); err != nil {
							return err
						}
						mapNode.Children[key] = nestedMap
					}

					if p.current().Type == TOKEN_DEDENT {
						p.advance()
					}
				}

				// Si el siguiente no es KEY, salir
				if p.current().Type != TOKEN_KEY {
					break
				}
			}

			node.Items = append(node.Items, mapNode)
		}

		if p.current().Type == TOKEN_DEDENT {
			break
		}
	}

	return nil
}

// ToMap convierte el AST a un mapa genérico de Go
func (n *Node) ToMap() map[string]interface{} {
	if n.Type != NODE_MAP {
		return nil
	}

	result := make(map[string]interface{})
	for key, child := range n.Children {
		switch child.Type {
		case NODE_STRING:
			result[key] = child.Value
		case NODE_MAP:
			result[key] = child.ToMap()
		case NODE_LIST:
			items := []interface{}{}
			for _, item := range child.Items {
				if item.Type == NODE_STRING {
					items = append(items, item.Value)
				} else if item.Type == NODE_MAP {
					items = append(items, item.ToMap())
				}
			}
			result[key] = items
		}
	}
	return result
}

func ParseYAML(content string) (*Node, error) {
	lexer := NewLexer(content)
	tokens := lexer.Tokenize()
	parser := NewParser(tokens)
	return parser.Parse()
}
