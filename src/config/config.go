package config

import "me.game/src/size"

type Config struct {
	Protocol string     //協定
	Port     int        //埠號
	World    int        //世界數量
	Split    int        //切割區域
	Size     size.Point //世界大小
}
