package main

import (
	"github.com/ActiveState/tail"
	"fmt"
	"strings"
	"strconv"
)


func main(){
	t, _ := tail.TailFile("/var/log/apache2/yyste_access.log", tail.Config{Follow: true})
	i := 0
	mResult := map[string]int{}
	for line := range t.Lines {
		arrResult := strings.Split( line.Text, " " )
		mResult[arrResult[8]] += 1
		i += 1
		if i%10 == 0 {
			fmt.Println( "i is "+strconv.Itoa(i) );
			for k,v := range mResult {
				fmt.Println( "code "+ k +":"+strconv.Itoa(v) );
			}
		}
	}
}


