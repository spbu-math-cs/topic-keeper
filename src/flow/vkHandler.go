package main

type VKListener struct {
	accessToken string
}

func (*VKListener) handleUpdates() {

}

func newVKListener() VKListener {
	return VKListener{}
}
