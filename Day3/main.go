package main

import (
	"fmt"
	"strings"
)

type Node[T any] struct {
	Data T
	Next *Node[T]
	Prev *Node[T]
	Cap  int
}

type DLinkedList[T any] struct {
	Head *Node[T]
	Tail *Node[T]
}

type FormatRange struct {
	Start  int
	Length int
	Bold   bool
	Italic bool
}

// Helper to safely copy format states
func cloneFormats(formats []FormatRange) []FormatRange {
	if formats == nil {
		return nil
	}
	cloned := make([]FormatRange, len(formats))
	copy(cloned, formats)
	return cloned
}

type EditOperation interface {
	Apply(doc *Document, Undo bool)
	String() string
}

type Operation struct {
	Position   int
	Action     string
	OldData    string // Data that was removed
	NewData    string // Data that was added
	Bold       bool
	Italic     bool
	OldFormats []FormatRange // Snapshot of formats before action
	NewFormats []FormatRange // Snapshot of formats after action
}

func (op *Operation) Apply(doc *Document, Undo bool) {
	if Undo {
		switch op.Action {
		case "insert":
			doc.deleteWithoutRecord(op.Position, len(op.NewData))
		case "delete":
			doc.insertWithoutRecord(op.Position, op.OldData)
		case "replace":
			doc.deleteWithoutRecord(op.Position, len(op.NewData))
			doc.insertWithoutRecord(op.Position, op.OldData)
		}
		// Restore exact formatting from before the action
		doc.Formats = cloneFormats(op.OldFormats)
	} else {
		switch op.Action {
		case "insert":
			doc.insertWithoutRecord(op.Position, op.NewData)
		case "delete":
			doc.deleteWithoutRecord(op.Position, len(op.OldData))
		case "replace":
			doc.deleteWithoutRecord(op.Position, len(op.OldData))
			doc.insertWithoutRecord(op.Position, op.NewData)
		}
		// Restore exact formatting from after the action
		doc.Formats = cloneFormats(op.NewFormats)
	}
}

func (op *Operation) String() string {
	var details string
	switch op.Action {
	case "insert":
		details = fmt.Sprintf("Added: '%s'", op.NewData)
	case "delete":
		details = fmt.Sprintf("Removed: '%s'", op.OldData)
	case "replace":
		details = fmt.Sprintf("Replaced: '%s' -> '%s'", op.OldData, op.NewData)
	case "format":
		details = fmt.Sprintf("Bold: %v, Italic: %v", op.Bold, op.Italic)
	}
	return fmt.Sprintf("Action: %s, Position: %d, %s", op.Action, op.Position, details)
}

type OperationNode struct {
	Op   EditOperation
	Next *OperationNode
	Prev *OperationNode
}

type OperationList struct {
	Head    *OperationNode
	Tail    *OperationNode
	Current *OperationNode
}

func NewOperationList() *OperationList {
	return &OperationList{}
}

func (ol *OperationList) Append(op EditOperation) {
	newNode := &OperationNode{Op: op}

	// If history was completely undone, start fresh
	if ol.Current == nil {
		ol.Head = newNode
		ol.Tail = newNode
		ol.Current = newNode
		return
	}

	// If we are in the middle of history, truncate the future
	if ol.Current != ol.Tail {
		ol.Current.Next = newNode
		newNode.Prev = ol.Current
		newNode.Next = nil
		ol.Tail = newNode
		ol.Current = newNode
		return
	}

	newNode.Prev = ol.Tail
	ol.Tail.Next = newNode
	ol.Tail = newNode
	ol.Current = newNode
}

func (ol *OperationList) String() string {
	var result string
	node := ol.Head
	index := 0
	for node != nil {
		marker := " "
		if node == ol.Current {
			marker = ">"
		}
		result += fmt.Sprintf("[%d]%s %s\n", index, marker, node.Op.String())
		node = node.Next
		index++
	}
	return result
}

type Document struct {
	Doc     *DLinkedList[string]
	History *OperationList
	Formats []FormatRange
}

func NewDocument() *Document {
	return &Document{
		Doc:     &DLinkedList[string]{},
		History: NewOperationList(),
	}
}

func (d *Document) addFormat(pos int, length int, bold bool, italic bool) {
	d.Formats = append(d.Formats, FormatRange{
		Start:  pos,
		Length: length,
		Bold:   bold,
		Italic: italic,
	})
}

func (d *Document) adjustFormatsForInsert(pos int, length int) {
	for i := range d.Formats {
		start := d.Formats[i].Start
		end := start + d.Formats[i].Length

		if pos <= start {
			d.Formats[i].Start += length
		} else if pos < end {
			d.Formats[i].Length += length
		}
	}
}

func (d *Document) adjustFormatsForDelete(pos int, length int) {
	newFormats := []FormatRange{}
	for _, f := range d.Formats {
		start := f.Start
		end := f.Start + f.Length
		delStart := pos
		delEnd := pos + length

		if end <= delStart {
			newFormats = append(newFormats, f)
			continue
		}
		if start >= delEnd {
			f.Start -= length
			newFormats = append(newFormats, f)
			continue
		}

		overlapStart := max(start, delStart)
		overlapEnd := min(end, delEnd)
		removed := overlapEnd - overlapStart
		f.Length -= removed

		if start > delStart {
			f.Start = delStart
		}
		if f.Length > 0 {
			newFormats = append(newFormats, f)
		}
	}
	d.Formats = newFormats
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (d *Document) insertWithoutRecord(position int, data string) {
	if d.Doc.Head == nil {
		node := &Node[string]{Data: data, Cap: 10}
		d.Doc.Head = node
		d.Doc.Tail = node
		return
	}

	remaining := position
	curr := d.Doc.Head

	for curr != nil && remaining > len(curr.Data) {
		remaining -= len(curr.Data)
		curr = curr.Next
	}

	if curr == nil {
		curr = d.Doc.Tail
		padding := remaining
		for padding > 0 {
			spaceLeft := curr.Cap - len(curr.Data)
			if spaceLeft == 0 {
				newNode := &Node[string]{Data: "", Cap: 10}
				d.Doc.insertAfter(curr, newNode)
				curr = newNode
				continue
			}
			add := padding
			if add > spaceLeft {
				add = spaceLeft
			}
			curr.Data += strings.Repeat(" ", add)
			padding -= add
		}
		remaining = len(curr.Data)
	}

	if remaining > len(curr.Data) {
		padding := remaining - len(curr.Data)
		curr.Data += strings.Repeat(" ", padding)
	}

	displacedData := curr.Data[remaining:]
	avail := curr.Cap - len(curr.Data)

	if avail >= len(data)+len(displacedData) {
		curr.Data = curr.Data[:remaining] + data + displacedData
		return
	}

	insertLen := avail
	if insertLen > len(data) {
		insertLen = len(data)
	}

	curr.Data = curr.Data[:remaining] + data[:insertLen]
	remainingInputData := data[insertLen:]
	remainingData := remainingInputData + displacedData

	newNode := &Node[string]{Data: remainingData, Cap: 10}
	d.Doc.insertAfter(curr, newNode)

	for len(newNode.Data) > newNode.Cap {
		overflow := newNode.Data[newNode.Cap:]
		newNode.Data = newNode.Data[:newNode.Cap]

		overflowNode := &Node[string]{Data: overflow, Cap: 10}
		d.Doc.insertAfter(newNode, overflowNode)
		newNode = overflowNode
	}
}

// Helper to clean up empty nodes after a deletion
func (d *Document) cleanEmptyNodes() {
	curr := d.Doc.Head
	for curr != nil {
		next := curr.Next
		if len(curr.Data) == 0 {
			if curr.Prev != nil {
				curr.Prev.Next = curr.Next
			} else {
				d.Doc.Head = curr.Next
			}
			if curr.Next != nil {
				curr.Next.Prev = curr.Prev
			} else {
				d.Doc.Tail = curr.Prev
			}
		}
		curr = next
	}
}

func (d *Document) deleteWithoutRecord(position int, length int) string {
	if d.Doc.Head == nil || length <= 0 {
		return ""
	}

	deletedData := ""
	remaining := position
	curr := d.Doc.Head

	for curr != nil && remaining > len(curr.Data) {
		remaining -= len(curr.Data)
		curr = curr.Next
	}

	if curr == nil {
		return ""
	}

	toDelete := length

	if remaining < len(curr.Data) {
		deleteLen := len(curr.Data) - remaining
		if deleteLen > toDelete {
			deleteLen = toDelete
		}
		deletedData += curr.Data[remaining : remaining+deleteLen]
		curr.Data = curr.Data[:remaining] + curr.Data[remaining+deleteLen:]
		toDelete -= deleteLen
	}

	next := curr.Next
	for toDelete > 0 && next != nil {
		if toDelete >= len(next.Data) {
			deletedData += next.Data
			toDelete -= len(next.Data)

			curr.Next = next.Next
			if next.Next != nil {
				next.Next.Prev = curr
			} else {
				d.Doc.Tail = curr
			}
			next = next.Next
		} else {
			deletedData += next.Data[:toDelete]
			next.Data = next.Data[toDelete:]
			toDelete = 0
		}
	}

	// Clean up any empty nodes left behind by deletion
	d.cleanEmptyNodes()
	return deletedData
}

func (d *Document) Insert(position int, data string) {
	oldFormats := cloneFormats(d.Formats)

	d.insertWithoutRecord(position, data)
	d.adjustFormatsForInsert(position, len(data))

	d.History.Append(&Operation{
		Position:   position,
		Action:     "insert",
		NewData:    data,
		OldFormats: oldFormats,
		NewFormats: cloneFormats(d.Formats),
	})
}

func (d *Document) Delete(position int, length int) {
	oldFormats := cloneFormats(d.Formats)

	deletedText := d.deleteWithoutRecord(position, length)
	d.adjustFormatsForDelete(position, length)

	d.History.Append(&Operation{
		Position:   position,
		Action:     "delete",
		OldData:    deletedText,
		OldFormats: oldFormats,
		NewFormats: cloneFormats(d.Formats),
	})
}

func (d *Document) Replace(position int, data string) {
	oldFormats := cloneFormats(d.Formats)

	deletedText := d.deleteWithoutRecord(position, len(data))
	d.adjustFormatsForDelete(position, len(data))
	d.insertWithoutRecord(position, data)
	d.adjustFormatsForInsert(position, len(data))

	d.History.Append(&Operation{
		Position:   position,
		Action:     "replace",
		OldData:    deletedText,
		NewData:    data,
		OldFormats: oldFormats,
		NewFormats: cloneFormats(d.Formats),
	})
}

func (d *Document) Format(position int, length int, bold bool, italic bool) {
	oldFormats := cloneFormats(d.Formats)

	d.addFormat(position, length, bold, italic)

	d.History.Append(&Operation{
		Position:   position,
		Action:     "format",
		Bold:       bold,
		Italic:     italic,
		OldFormats: oldFormats,
		NewFormats: cloneFormats(d.Formats),
	})
}

func (d *Document) Undo() bool {
	if d.History.Current == nil {
		return false
	}

	// Apply undo for the current operation
	d.History.Current.Op.Apply(d, true)
	// Move the pointer back
	d.History.Current = d.History.Current.Prev
	return true
}

func (d *Document) Redo() bool {
	var nextOp *OperationNode
	if d.History.Current == nil {
		nextOp = d.History.Head
	} else {
		nextOp = d.History.Current.Next
	}

	if nextOp == nil {
		return false
	}

	// Apply redo for the next operation
	nextOp.Op.Apply(d, false)
	// Move the pointer forward
	d.History.Current = nextOp
	return true
}

func (d *Document) getFormatAt(pos int) (bool, bool) {
	bold := false
	italic := false
	for _, f := range d.Formats {
		if pos >= f.Start && pos < f.Start+f.Length {
			if f.Bold {
				bold = true
			}
			if f.Italic {
				italic = true
			}
		}
	}
	return bold, italic
}

func (d *Document) DisplayDocument() string {
	text := ""
	curr := d.Doc.Head
	for curr != nil {
		text += curr.Data
		curr = curr.Next
	}

	result := ""
	prevBold := false
	prevItalic := false

	for i, ch := range text {
		bold, italic := d.getFormatAt(i)

		if bold && !prevBold {
			result += "[B]"
		}
		if italic && !prevItalic {
			result += "[I]"
		}
		if !bold && prevBold {
			result += "[/B]"
		}
		if !italic && prevItalic {
			result += "[/I]"
		}

		result += string(ch)

		prevBold = bold
		prevItalic = italic
	}

	if prevBold {
		result += "[/B]"
	}
	if prevItalic {
		result += "[/I]"
	}

	return result
}

func (d *Document) DisplayDocumentWithNodes() string {
	var result string
	result += "Document Content: " + d.DisplayDocument() + "\n\n"
	result += "Node Structure:\n"
	result += "==============\n"

	curr := d.Doc.Head
	nodeNum := 1

	for curr != nil {
		result += fmt.Sprintf(
			"Node %d: Data=\"%s\", Len=%d, Cap=%d\n",
			nodeNum,
			curr.Data,
			len(curr.Data),
			curr.Cap,
		)
		curr = curr.Next
		nodeNum++
	}

	if d.Doc.Head == nil {
		result += "[Empty Document]\n"
	}

	return result
}

func (d *Document) DisplayHistory() string {
	return d.History.String()
}

func (l *DLinkedList[T]) insertAfter(node *Node[T], newNode *Node[T]) {
	newNode.Prev = node
	newNode.Next = node.Next

	if node.Next != nil {
		node.Next.Prev = newNode
	} else {
		l.Tail = newNode
	}
	node.Next = newNode
}

func main() {
	myDocument := NewDocument()
	var userChoice int

	for {
		fmt.Println("\n--------------------------------------------------")
		fmt.Println("Enter 1 for Insert")
		fmt.Println("Enter 2 for Delete")
		fmt.Println("Enter 3 for Replace")
		fmt.Println("Enter 4 for Undo")
		fmt.Println("Enter 5 for Redo")
		fmt.Println("Enter 6 for Display Document")
		fmt.Println("Enter 7 for Display History")
		fmt.Println("Enter 8 for Format")
		fmt.Println("Enter 0 for Exit")
		fmt.Println("--------------------------------------------------")

		fmt.Print("Choice: ")
		fmt.Scan(&userChoice)

		switch userChoice {
		case 1:
			var position int
			var data string
			fmt.Print("Enter position to insert: ")
			fmt.Scan(&position)
			fmt.Print("Enter data to insert: ")
			fmt.Scan(&data)
			myDocument.Insert(position, data)
			fmt.Println("Inserted successfully!")

		case 2:
			var position, length int
			fmt.Print("Enter position to delete: ")
			fmt.Scan(&position)
			fmt.Print("Enter length to delete: ")
			fmt.Scan(&length)
			myDocument.Delete(position, length)
			fmt.Println("Deleted successfully!")

		case 3:
			var position int
			var data string
			fmt.Print("Enter position to replace: ")
			fmt.Scan(&position)
			fmt.Print("Enter replacement text: ")
			fmt.Scan(&data)
			myDocument.Replace(position, data)
			fmt.Println("Replaced successfully!")

		case 4:
			if myDocument.Undo() {
				fmt.Println("✓ Undo successful!")
			} else {
				fmt.Println("✗ Cannot undo!")
			}

		case 5:
			if myDocument.Redo() {
				fmt.Println("✓ Redo successful!")
			} else {
				fmt.Println("✗ Cannot redo!")
			}

		case 6:
			fmt.Println("\n" + myDocument.DisplayDocumentWithNodes())

		case 7:
			fmt.Println("\nOperation History:")
			fmt.Println(myDocument.DisplayHistory())

		case 8:
			var position int
			var length int
			var boldInput string
			var italicInput string

			fmt.Print("Enter position: ")
			fmt.Scan(&position)
			fmt.Print("Enter length: ")
			fmt.Scan(&length)
			fmt.Print("Bold (y/n): ")
			fmt.Scan(&boldInput)
			fmt.Print("Italic (y/n): ")
			fmt.Scan(&italicInput)

			bold := boldInput == "y"
			italic := italicInput == "y"

			myDocument.Format(position, length, bold, italic)
			fmt.Println("Format applied!")

		case 0:
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Invalid choice!")
		}
	}
}
