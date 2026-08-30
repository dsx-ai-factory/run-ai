package enforcer

import (
	"context"
	"encoding/json"
	jsonq "github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/conjungo"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

//		 before merging the values, we "sign" each field with the index of the level it resided in.
//		 we need this in order to know from which level each value in the result was produced. example:
//			levelElement[0]
//				value-for-gpu 		-> sign with 0
//			levelElement[1]
//				default-for-image	-> sign with 1
//				value-for-gpu		-> sign with 1
//		 the end result after merge all layers:
//			value-for-gpu --> 0: final value obtained from the workload
//			value-for-image --> 1: final value obtained from the policy (workload did not override it)
//
//	  note that for itemized fields, each item is signed individually, because each item can be produced
//	  from different level. for example one volume can be derived from the policy defaults, and
//	  another volume can be provided by the researcher in the workload.
//
//	  note that for itemized fields we sign not only vlaues but also "deletion" flags, cause we want to
//	  know in which level the item has been deleted. was it deleted in the base workload and then nee value
//	  provided for it in the topmost workload? or maybe there was a value for it in the base workload which
//	  our topmost workload deleted.
func (self *MergeStack) signEachLevel(_ context.Context) {
	//
	//   for each level...
	//
	for levelIdx, levelElement := range self.levelElement {

		//
		//   we want only the type of the level desc, which is the first token in its desc
		//
		levelType := strings.Split(self.levelDesc[levelIdx], ":")[0]

		//
		//   for each fields under "spec"...
		//
		levelFlds := levelElement.JumpToTop()
		for fldName := range utils.MapMe(levelFlds.Get()) {

			fldDefMap := utils.MapMe(levelFlds.JumpToTop().Find(fldName))

			//
			//   sign any primitive value
			//
			if fldDefMap[common.Value] != nil {
				fldDefMap[common.ValueFromLevel] = levelIdx
				fldDefMap[common.Level] = levelType
			}

			//
			//    do the signing for inner members of items and values
			//
			for _, clause := range []string{common.Items, common.Values} {
				if fldDefMap[clause] != nil {
					//
					//   if this is a map containing items, then for each item add the valueFrom clause
					//
					itemsMap := utils.MapMe(fldDefMap[clause])
					for _, itemDef := range itemsMap {
						itemDefMap := utils.MapMe(itemDef)
						//
						//   in the same level - value should either be deleted or have a value. if both value and
						//   deletion flag co-exist in the same level we give precedence to deletion.
						//
						if itemDefMap[common.Deleted] != nil && utils.BoolMe(itemDefMap[common.Deleted], false) {
							itemDefMap[common.DeletedLevel] = levelIdx
							itemDefMap[common.Level] = levelType
							delete(itemDefMap, common.Value)
						}
						if itemDefMap[common.Value] != nil {
							itemDefMap[common.ValueFromLevel] = levelIdx
							itemDefMap[common.Level] = levelType
						}
					}
				}
			}
		}
	}
}

//	  merging is done by third party conjungo library. we're doing it from the bottom layer up
//	  if we have 3 levels (2, 1, 0) then we merge:
//			workload 0 -> workload 1, result goes to level 1
//		    workload 1 -> policy 2, result goes to level 2
//	  so the final result always goes to the lowest element of the stack, containing the policy
func (self *MergeStack) merge(ctx context.Context) error {
	log := logger.FromContext(ctx)
	var lastError error

	for level := 1; level < len(self.levelElement); level += 1 {
		target := self.levelElement[level].JumpToTop().Get()
		source := self.levelElement[level-1].JumpToTop().Get()
		err := conjungo.Merge(&target, source, conjungo.NewOptions())
		if err != nil {
			lastError = err
			log.Error(err, "error occurred during merge",
				"of", self.levelDesc[level-1], "to", self.levelDesc[level])
		}
	}

	return lastError
}

// return the lowest level of the stack, where the final merged values can be found
func (self *MergeStack) MergedValues() *jsonq.JSONQ {
	return self.levelElement[len(self.levelElement)-1]
}

func (self *MergeStack) PolicyLevel() int {
	return len(self.levelElement) - 1
}
func (self *MergeStack) MergedValuesLevel() int {
	return len(self.levelElement) - 1
}

// return the level on top of the policy. this would contain the merged values
// from all the levels excluding the policy level
// 3 levels (workload + base + policy) will return 1
// 2 levels (workload + policy) will return 0
func (self *MergeStack) MergedValuesExcludingPolicy() int {
	return len(self.levelElement) - 2
}

// do some cleanup of the workload after merge:
//   - remove rules
//   - remove value-from and deleted-from signing
//   - remove deletion flag from non-deleted items
//
// returned the clean workload from the stack
type CleanOptions struct {
	KeepLevels bool // do not clean the "level" field
	KeepRules  bool
}

func (self *MergeStack) CleanTheWorkload(level int, options CleanOptions) *jsonq.JSONQ {

	if level < 0 || level >= len(self.levelElement) {
		return nil
	}

	mergedValues := self.levelElement[level].JumpToTop()
	if mergedValues.Error() != nil {
		return self.levelElement[level]
	}

	//
	//   go over the fields
	//
	for fldName := range utils.MapMe(mergedValues.Get()) {

		mergedFldDef := mergedValues.JumpTo(fldName)
		mergedFldMap := utils.MapMe(mergedFldDef.JumpToTop().Get())

		//
		//   remove the rules
		//
		if !options.KeepRules {
			delete(mergedFldMap, common.KeyRules)
			delete(mergedFldMap, common.ItemRules)
			delete(mergedFldMap, common.Rules)
		}

		delete(mergedFldMap, common.ValueFromLevel)
		if !options.KeepLevels {
			delete(mergedFldMap, common.Level)
		}

		//
		//   for itemized field, go over each item/member
		//
		for _, clause := range []string{common.Items, common.Values} {

			mergedFldItems := mergedFldDef.JumpTo(clause)

			for itemName := range utils.MapMe(mergedFldItems.Get()) {

				thisItemDef := mergedFldItems.JumpTo(itemName)
				thisItemMap := utils.MapMe(thisItemDef.JumpToTop().Get())

				//
				//    delete rules but keep
				//
				if !options.KeepRules {
					delete(thisItemMap, common.Rules)
				}
				if !options.KeepLevels {
					delete(thisItemMap, common.Level)
				}

				//
				//    case item has both deletion flag and value need to check based on the
				//    levels. if it has been deleted - remove the value. if it has not been
				//    deleted - remove the deletion flag leave only the value
				//
				deletionLevel := thisItemDef.JumpToTop().Find(common.DeletedLevel)
				valueFromLevel := thisItemDef.JumpToTop().Find(common.ValueFromLevel)

				if deletionLevel != nil && valueFromLevel != nil {
					if deletionLevel.(int) < valueFromLevel.(int) { // lower level wins
						delete(thisItemMap, common.Value)
					} else {
						delete(thisItemMap, common.Deleted)
					}
				}

				delete(thisItemMap, common.ValueFromLevel)
				delete(thisItemMap, common.DeletedLevel)
			}
		}
	}

	return self.levelElement[level]
}

// this function is tailored for the frontend needs, and it needs to be used only for this purpose.
// it is triggered by the webhook, when "merged" flag is sent with value. it is always rejected with
// error, and the error is whatever we return from this function: a json containing the complete
// merged workload including runai rules. we clean it from attributes which are not relevant for the
// client, but leave the rules and level attributes cause the frontend needs them.
func (self *MergeStack) MergeWorkloadForFrontend() ([]byte, error) {

	self.CleanTheWorkload(self.MergedValuesLevel(), CleanOptions{
		KeepLevels: true,
		KeepRules:  true,
	})

	runaiRules := utils.MapMe(self.runaiRules.Get())
	err := conjungo.Merge(&runaiRules, self.MergedValues().JumpToTop().Get(), conjungo.NewOptions())
	if err != nil {
		return nil, err
	}

	return json.Marshal(self.runaiRules.Get())
}
