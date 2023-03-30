package listen

import (
	"github.com/gogf/gf/v2/util/gconv"
)

func HasMember(tar interface{}, sli interface{}) (idx int, ok bool) {
	slis := gconv.SliceAny(sli)

	for idx, ori := range slis {
		if tar == ori {
			return idx, true
		}
	}

	return 0, false
}

func IsStrInList(str string, ls []string) (idx int, has bool) {
	for idx, l := range ls {
		if l == str {
			return idx, true
		}
	}

	return -1, false
}
