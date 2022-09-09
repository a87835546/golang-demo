package error_utils

type ServiceErrorModel struct {
	Code int32
	Msg  string
}

func (model *ServiceErrorModel) Error() string {
	return model.Error()
}

func (model *ServiceErrorModel) buildErrorModel() string {
	return model.Error()
}
