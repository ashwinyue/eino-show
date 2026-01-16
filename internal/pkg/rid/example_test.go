package rid_test

import (
	"fmt"

	"github.com/ashwinyue/eino-show/internal/pkg/rid"
)

func ExampleResourceID_String() {
	// 定义一个资源标识符，例如用户资源
	userID := rid.UserID

	// 调用 String 方法，将 ResourceID 类型转换为字符串类型
	idString := userID.String()

	// 输出结果
	fmt.Println(idString)

	// Output:
	// user
}
