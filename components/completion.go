package components

import (
	"context"
	"pls/proto/view"

	"github.com/TobiasYin/go-lsp/lsp/defines"
)

var protoKeywordCompletionItems []defines.CompletionItem

func init() {
	kindKeyword := defines.CompletionItemKindKeyword
	for _, keyword := range []string{"string", "bytes", "double", "float", "int32", "int64",
		"uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool",
		"message", "enum", "service", "rpc", "optional", "repeated", "required",
		"option", "default", "syntax", "package", "import", "extend", "oneof", "map", "reserved",
	} {
		insertText := keyword
		protoKeywordCompletionItems = append(protoKeywordCompletionItems, defines.CompletionItem{
			Kind:       &kindKeyword,
			Label:      keyword,
			InsertText: &insertText,
		})
	}
}

func Completion(ctx context.Context, req *defines.CompletionParams) (*[]defines.CompletionItem, error) {
	if !view.IsProtoFile(req.TextDocument.Uri) {
		return nil, nil
	}
	proto_file, err := view.ViewManager.GetFile(req.TextDocument.Uri)
	if err != nil || proto_file.Proto() == nil {
		return nil, nil
	}
	line_str := proto_file.ReadLine(int(req.Position.Line))
	word := getWord(line_str, int(req.Position.Character-1), false)
	wordWithDot := getWord(line_str, int(req.Position.Character-1), true)
	if req.Context.TriggerKind != defines.CompletionTriggerKindTriggerCharacter {
		res, err := CompletionInThisFile(proto_file)
		*res = append(*res, GetImportedPackages(proto_file)...)

		return res, err
	}
	for _, im := range proto_file.Proto().Imports() {
		import_uri, err := view.GetDocumentUriFromImportPath(req.TextDocument.Uri, im.ProtoImport.Filename)
		if err != nil {
			continue
		}

		file, err := view.ViewManager.GetFile(import_uri)
		if err != nil {
			continue
		}

		if len(file.Proto().Packages()) == 0 {
			continue
		}

		packageName := file.Proto().Packages()[0].ProtoPackage.Name
		if packageName == word || packageName+"." == wordWithDot {
			return CompletionInThisFile(file)
		}
	}
	return nil, nil
}

func GetImportedPackages(proto_file view.ProtoFile) (res []defines.CompletionItem) {
	unique := make(map[string]struct{})
	for _, im := range proto_file.Proto().Imports() {
		import_uri, err := view.GetDocumentUriFromImportPath(proto_file.URI(), im.ProtoImport.Filename)
		if err != nil {
			continue
		}

		file, err := view.ViewManager.GetFile(import_uri)
		if err != nil {
			continue
		}

		if len(file.Proto().Packages()) == 0 {
			continue
		}

		packageName := file.Proto().Packages()[0].ProtoPackage.Name

		if _, exist := unique[packageName]; exist {
			continue
		}

		kindModule := defines.CompletionItemKindModule
		unique[packageName] = struct{}{}
		res = append(res, defines.CompletionItem{
			Label:      packageName,
			Kind:       &kindModule,
			InsertText: &packageName,
		})

	}

	return res
}

func CompletionInThisFile(file view.ProtoFile) (result *[]defines.CompletionItem, err error) {
	kindEnum := defines.CompletionItemKindEnum
	res := protoKeywordCompletionItems
	for _, enums := range file.Proto().Enums() {

		doc := formatHover(SymbolDefinition{
			Type: DefinitionTypeEnum,
			Enum: enums,
		})

		res = append(res, defines.CompletionItem{
			Label:      enums.Protobuf().Name,
			Kind:       &kindEnum,
			InsertText: &enums.Protobuf().Name,
			Documentation: defines.MarkupContent{
				Kind:  defines.MarkupKindMarkdown,
				Value: doc,
			},
		})
	}

	kindClass := defines.CompletionItemKindClass
	for _, message := range file.Proto().Messages() {

		doc := formatHover(SymbolDefinition{
			Type:    DefinitionTypeMessage,
			Message: message,
		})

		res = append(res, defines.CompletionItem{
			Label:      message.Protobuf().Name,
			Kind:       &kindClass,
			InsertText: &message.Protobuf().Name,
			Documentation: defines.MarkupContent{
				Kind:  defines.MarkupKindMarkdown,
				Value: doc,
			},
		})
	}
	return &res, nil
}
