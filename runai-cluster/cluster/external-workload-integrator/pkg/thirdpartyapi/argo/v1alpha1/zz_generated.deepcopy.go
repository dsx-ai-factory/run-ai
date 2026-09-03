// Hand-written deepcopy methods for the slim Workflow stand-in. The file is
// named zz_generated for parity with controller-gen output, but it is no
// longer machine-generated: we only carry the few types the handler reads.
package v1alpha1

func (in *Workflow) DeepCopyInto(out *Workflow) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *Workflow) DeepCopy() *Workflow {
	if in == nil {
		return nil
	}
	out := new(Workflow)
	in.DeepCopyInto(out)
	return out
}

func (in *WorkflowList) DeepCopyInto(out *WorkflowList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Workflow, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *WorkflowList) DeepCopy() *WorkflowList {
	if in == nil {
		return nil
	}
	out := new(WorkflowList)
	in.DeepCopyInto(out)
	return out
}

func (in *WorkflowSpec) DeepCopyInto(out *WorkflowSpec) {
	*out = *in
	if in.Templates != nil {
		out.Templates = make([]Template, len(in.Templates))
		for i := range in.Templates {
			in.Templates[i].DeepCopyInto(&out.Templates[i])
		}
	}
}

func (in *Template) DeepCopyInto(out *Template) {
	*out = *in
	if in.Container != nil {
		out.Container = in.Container.DeepCopy()
	}
}

func (in *WorkflowStatus) DeepCopyInto(out *WorkflowStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make(Conditions, len(in.Conditions))
		copy(out.Conditions, in.Conditions)
	}
}
