package core

import (
	"github.com/seapvnk/qokl/storage"
)

// Entity module setup
func (vm *VM) UseEntityModule() *VM {
	vm.environment.AddFunction("insert", storage.FnEntityInsert)
	vm.environment.AddFunction("deleteEntity", storage.FnDeleteEntity)
	vm.environment.AddFunction("entity", storage.FnEntityGet)
	vm.environment.AddFunction("select", storage.FnEntitySelect)
	vm.environment.AddFunction("addTag", storage.FnAddTag)
	vm.environment.AddFunction("relationship", storage.FnRelationship)
	vm.environment.AddFunction("relationshipsOf", storage.FnEntityRelationships)

	return vm
}
