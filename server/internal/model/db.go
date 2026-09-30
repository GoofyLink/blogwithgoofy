package model

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"blogwitgoofy/server/internal/config"
	"blogwitgoofy/server/pkg/utils"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port,
		cfg.MySQL.Database, cfg.MySQL.Charset)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		log.Fatalf("连接 MySQL 失败（请确认 phpstudy 已启动 MySQL 且 %s 库已创建）: %v", cfg.MySQL.Database, err)
	}
	DB = db

	// 连接池：防止复用被 MySQL 掐断的死连接（表现为偶发 invalid connection）。
	// 注意：本机 phpstudy 的 wait_timeout=120s（连接空闲 2 分钟即被服务端关闭），
	// 因此空闲回收时限必须小于它。
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetConnMaxIdleTime(90 * time.Second)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetMaxOpenConns(30)
	}

	if err := DB.AutoMigrate(
		&User{}, &Category{}, &Tag{}, &Article{},
		&Annotation{}, &Comment{}, &Link{}, &Page{},
		&Book{}, &Chapter{}, &ChapterNote{}, &Word{}, &Setting{}, &Anime{}, &Artwork{}, &AiTool{}, &AiPrompt{}, &Game{}, &ApiLog{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	seed(db)
}

func seed(db *gorm.DB) {
	// 管理员
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		hash, _ := utils.HashPassword("admin123")
		db.Create(&User{Username: "admin", PasswordHash: hash, Nickname: "管理员"})
		log.Println("已创建默认管理员 admin / admin123，登录后请尽快修改密码")
	}

	// 默认分类
	var catCount int64
	db.Model(&Category{}).Count(&catCount)
	if catCount == 0 {
		db.Create(&[]Category{{Name: "默认分类"}, {Name: "技术笔记"}, {Name: "生活随笔"}})
	}

	// 关于页
	var pageCount int64
	db.Model(&Page{}).Count(&pageCount)
	if pageCount == 0 {
		db.Create(&Page{
			Slug:  "about",
			Title: "关于我",
			Content: "## 你好，欢迎来到我的博客\n\n这里用 Vue3 + Go 写成。\n\n" +
				"- 📝 记录技术与生活\n- 🇬🇧 🇩🇪 顺便学英语和德语\n\n欢迎常来逛逛！",
		})
	}

	// 示例文章（含学习文章 + 批注演示）
	var articleCount int64
	db.Model(&Article{}).Count(&articleCount)
	if articleCount == 0 {
		var defaultCat Category
		db.Where("name = ?", "默认分类").First(&defaultCat)

		blogPost := Article{
			Title: "Hello World：我的第一篇博客",
			Summary: "博客正式开张！简单介绍这个站是怎么搭起来的：Vue3 + Vite 前台、" +
				"Gin + GORM 后端、MySQL 存储，外加一个英德语学习模块。",
			Content: "## 为什么写博客\n\n把学过的东西写下来，才算真的学会。这个站用来记录技术笔记和生活随笔。" +
				"\n\n## 技术栈\n\n- 前端：Vue3 + TypeScript + Vite + Element Plus\n" +
				"- 后端：Go + Gin + GORM\n- 数据库：MySQL\n\n" +
				"## 学习模块\n\n除了写博客，站里还有一个「学习」板块，可以边读英语/德语文章边加批注。\n\n" +
				"```go\nfunc main() {\n    fmt.Println(\"Hello, Blog!\")\n}\n```\n\n开写！",
			CategoryID: defaultCat.ID,
			Type:       ArticleTypeBlog,
			Language:   "zh",
			Status:     StatusPublished,
		}
		db.Create(&blogPost)

		learnEn := Article{
			Title: "Why Reading Aloud Helps You Learn English",
			Summary: "An easy English article about the benefits of reading aloud. " +
				"试试选中段落里的文字，然后点击「添加批注」写下你的学习笔记。",
			Content: "Reading aloud is one of the oldest and most effective ways to learn a language. " +
				"When you speak the words yourself, your brain processes them twice: once through your eyes and once through your mouth.\n\n" +
				"Many learners read silently because it is faster. However, silent reading often hides pronunciation problems. " +
				"You may know a word when you see it, but you cannot say it confidently in a conversation.\n\n" +
				"Reading aloud also builds rhythm. English is a stress-timed language, which means some syllables are long and loud while others are short and quiet. " +
				"Practicing out loud trains your mouth to follow this natural music.\n\n" +
				"Finally, reading aloud improves memory. Words that you have spoken are easier to remember than words you have only seen. " +
				"Try it with this article: read one paragraph aloud every day, and add your own notes beside the text.",
			Translation: "朗读是学习语言最古老、最有效的方法之一。当你自己把词说出来，大脑会对它们进行两次加工：一次通过眼睛，一次通过嘴巴。\n\n" +
				"很多学习者默读，因为这样更快。然而默读常常掩盖发音问题。你看到某个词时可能认识它，但在对话中却不敢自信地说出来。\n\n" +
				"朗读还能培养节奏感。英语是重音节拍语言，有些音节又长又响，有些又短又轻。出声练习能让你的嘴巴跟上这种天然的韵律。\n\n" +
				"最后，朗读能增强记忆。你说过的词比你只看过的词更容易记住。用这篇文章试试：每天朗读一个段落，并在旁边写下你自己的笔记。",
			CategoryID: defaultCat.ID,
			Type:       ArticleTypeLearning,
			Language:   "en",
			Status:     StatusPublished,
		}
		db.Create(&learnEn)

		db.Create(&[]Annotation{
			{ArticleID: learnEn.ID, ParagraphIndex: 0, Quote: "your brain processes them twice",
				Note: "process 加工/处理。这里 twice 指眼睛+嘴巴双通道，很好的表达，写作文能用上。"},
			{ArticleID: learnEn.ID, ParagraphIndex: 2, Quote: "stress-timed language",
				Note: "重音节拍语言。英语按重音划分节奏，和汉语的音节节拍不同，这是听不懂快速英语的重要原因。"},
		})

		learnDe := Article{
			Title: "Warum Deutsch lernen lohnt sich",
			Summary: "Ein einfacher deutscher Text über die Vorteile des Deutschlernens. 一篇简单的德语短文。",
			Content: "Deutsch ist die meistgesprochene Muttersprache in Europa. Über hundert Millionen Menschen sprechen Deutsch als erste oder zweite Sprache.\n\n" +
				"Deutsch klingt vielleicht schwierig, aber die Grammatik folgt klaren Regeln. Wenn man die Regeln einmal verstanden hat, kann man viele neue Wörter selbst bilden.\n\n" +
				"Zum Beispiel: das Wort \"Sprache\" bedeutet language. Mit der Endung \"-lich\" wird daraus \"sprachlich\" – das bedeutet linguistic. Deutsch baut große Wörter aus kleinen Teilen.\n\n" +
				"Wer Deutsch lernt, kann nicht nur in Deutschland, Österreich und der Schweiz arbeiten, sondern auch die Kultur von Goethe und Schiller im Original entdecken.",
			Translation: "德语是欧洲使用人数最多的母语。超过一亿人把德语作为第一或第二语言。\n\n" +
				"德语听起来也许很难，但语法遵循清晰的规则。一旦理解了这些规则，就能自己构造许多新词。\n\n" +
				"例如：\"Sprache\" 这个词的意思是语言。加上后缀 \"-lich\" 就变成 \"sprachlich\"，意思是「语言的」。德语用小零件搭建大单词。\n\n" +
				"学德语的人不仅可以在德国、奥地利和瑞士工作，还能用原文感受歌德和席勒的文化。",
			CategoryID: defaultCat.ID,
			Type:       ArticleTypeLearning,
			Language:   "de",
			Status:     StatusPublished,
		}
		db.Create(&learnDe)
	}

	// 学习模块书籍（公版内容：伊索寓言 / 布雷姆音乐家）
	var bookCount int64
	db.Model(&Book{}).Count(&bookCount)
	if bookCount == 0 {
		aesop := Book{
			Title:    "伊索寓言精选",
			Subtitle: "Aesop's Fables",
			Author:   "Aesop",
			Language: LangEN,
			Description: "经典伊索寓言选读，短句多、生词少，适合英语初级学习者逐句精读。" +
				"阅读时把鼠标移到单词上看释义，点击句子查看整句翻译。",
			Status: 1,
			Sort:    1,
		}
		db.Create(&aesop)

		db.Create(&[]Chapter{
			{
				BookID: aesop.ID, Sort: 1, Status: 1,
				Title: "Chapter 1 · The Ant and the Grasshopper",
				Content: "In a field one summer's day a Grasshopper was hopping about, chirping and singing to its heart's content. " +
					"An Ant passed by, bearing along with great effort an ear of corn he was taking to the nest.\n\n" +
					"\"Why not come and chat with me,\" said the Grasshopper, \"instead of toiling in that way?\"\n\n" +
					"\"I am helping to lay up food for the winter,\" said the Ant, \"and recommend you to do the same.\"\n\n" +
					"\"Why bother about winter?\" said the Grasshopper; \"we have got plenty of food at present.\"\n\n" +
					"But the Ant went on its way and continued its toil. When the winter came the Grasshopper had no food, " +
					"and found itself dying of hunger, while it saw the ants distributing every day corn and grain from the stores they had collected in the summer. " +
					"Then the Grasshopper knew: it is best to prepare for the days of necessity.",
				Translation: "夏日的一天，一只蚱蜢在田野里跳来跳去，尽情地鸣唱。一只蚂蚁经过，正吃力地把一穗玉米拖回巢穴。\n\n" +
					"“为什么不过来和我聊聊呢？”蚱蜢说，“何必这样辛苦劳作？”\n\n" +
					"“我在为冬天储备粮食，”蚂蚁说，“建议你也这样做。”\n\n" +
					"“何必操心冬天？”蚱蜢说，“我们现在有的是食物。”\n\n" +
					"但蚂蚁继续赶路劳作。冬天来临，蚱蜢没有食物，眼看就要饿死，而它看到蚂蚁们每天都在分发夏天储存的粮食。这时蚱蜢明白了：未雨绸缪才是上策。",
			},
			{
				BookID: aesop.ID, Sort: 2, Status: 1,
				Title: "Chapter 2 · The Boy Who Cried Wolf",
				Content: "There was once a young Shepherd Boy who tended his sheep at the foot of a mountain near a dark forest. " +
					"It was rather lonely for him all day, so he thought upon a plan by which he could get a little company and some excitement. " +
					"He rushed down towards the village calling out \"Wolf! Wolf!\"\n\n" +
					"The villagers came out to meet him, and some of them stopped with him for a considerable time. " +
					"This pleased the boy so much that a few days afterwards he tried the same trick, and again the villagers came to his help.\n\n" +
					"But shortly after this a Wolf actually did come out from the forest. The boy ran towards the village shouting \"Wolf! Wolf!\" " +
					"as loudly as he could. But the villagers thought he was trying to fool them again, so no one paid any attention to him. " +
					"The Wolf killed a great many of the boy's sheep and then slipped away into the forest.\n\n" +
					"Liar will not be believed, even when he speaks the truth.",
				Translation: "从前有一个牧童，在森林附近的山脚下放羊。他整天很孤单，于是想出一个能找人作伴、找点乐子的主意。他冲下山坡朝村子大喊：“狼来了！狼来了！”\n\n" +
					"村民们出来迎接他，有些人还陪了他好一阵子。男孩非常高兴，几天后又故技重施，村民们又来帮忙。\n\n" +
					"但不久之后，狼真的从森林里出来了。男孩拼命朝村子奔跑，大声喊着：“狼来了！狼来了！”但村民们以为他又在骗人，谁也不理他。狼咬死了男孩的许多羊，然后溜回了森林。\n\n" +
					"说谎的人即使说了真话，也不会有人相信。",
			},
			{
				BookID: aesop.ID, Sort: 3, Status: 1,
				Title: "Chapter 3 · The Tortoise and the Hare",
				Content: "A Hare was making fun of the Tortoise one day for being so slow. \"Do you ever get anywhere?\" he asked with a mocking laugh.\n\n" +
					"\"Yes,\" replied the Tortoise, \"and I get there sooner than you think. I'll run you a race and prove it.\"\n\n" +
					"The Hare was much amused at the idea of running a race with the Tortoise, but for the fun of the thing he agreed. " +
					"So the Fox, who had consented to act as judge, marked the distance and started the runners off.\n\n" +
					"The Hare was soon far out of sight, and to make the Tortoise feel very deeply how ridiculous it was for him to try a race with a Hare, " +
					"he lay down beside the course to take a nap until the Tortoise should catch up.\n\n" +
					"The Tortoise meanwhile kept going slowly but steadily, and, after a time, passed the place where the Hare was sleeping. " +
					"But the Hare slept on very peacefully; and when at last he did wake, the Tortoise was near the goal. The Hare now ran his swiftest, " +
					"but he could not overtake the Tortoise in time for the race was won by the slow but steady animal.\n\n" +
					"Slow but steady wins the race.",
				Translation: "一天，兔子嘲笑乌龟爬得太慢。“你到底能不能到得了终点？”他嘲笑着问。\n\n" +
					"“能，”乌龟回答，“而且到得比你想象的快。我跟你赛跑一场，证明给你看。”\n\n" +
					"兔子觉得和乌龟赛跑的想法很好笑，但为了好玩还是答应了。于是请狐狸做裁判，量好距离，比赛开始。\n\n" +
					"兔子很快跑得无影无踪。为了狠狠羞辱乌龟竟敢和自己赛跑，他在赛道边躺下睡了一觉，想等乌龟追上来。\n\n" +
					"乌龟则缓慢而坚定地前进，一段时间后爬过了兔子睡觉的地方。兔子睡得很沉，等他终于醒来，乌龟已经接近终点。兔子拼命狂奔，但还是来不及——缓慢而坚定的乌龟赢得了比赛。\n\n" +
					"稳扎稳打，方能制胜。",
			},
		})

		bremen := Book{
			Title:    "布雷姆的音乐家",
			Subtitle: "Die Bremer Stadtmusikanten",
			Author:   "Brüder Grimm",
			Language: LangDE,
			Description: "格林童话《不来梅的音乐家》简写版，常用德语词汇和句型，适合德语入门阅读。" +
				"鼠标悬停查词，点击句子翻译。",
			Status: 1,
			Sort:    2,
		}
		db.Create(&bremen)

		db.Create(&[]Chapter{
			{
				BookID: bremen.ID, Sort: 1, Status: 1,
				Title: "Kapitel 1 · Ein Esel auf Reisen 驴子出发",
				Content: "Ein Esel hatte lange Jahre getreue Dienste geleistet, aber jetzt war er alt und konnte nicht mehr arbeiten. " +
					"Sein Herr wollte ihn nicht mehr behalten, also lief der Esel weg nach Bremen. Dort, dachte er, konnte er Stadtmusikant werden.\n\n" +
					"Als er eine Weile gegangen war, fand er einen Jagdhund auf dem Weg, der lag und keuchte wie einer, der sich müde gelaufen hatte. " +
					"\"Warum keuchst du denn so, Bruder?\" fragte der Esel. \"Ach\", sagte der Hund, \"weil ich alt bin und jeden Tag jagen muss, " +
					"kann mein Herr mich nicht mehr brauchen und hat mich weggejagt.\"\n\n" +
					"\"Weißt du was?\" sprach der Esel, \"ich gehe nach Bremen und werde Stadtmusikant. Geh mit mir und lass dich auch in der Musik annehmen. " +
					"Ich spiele die Laute, und du schlägst die Pauken.\" Der Hund war zufrieden, und sie gingen weiter ihres Weges.\n\n" +
					"Nicht lange danach saß eine Katze auf dem Weg und machte ein Gesicht wie drei Tage Regenwetter. " +
					"\"Was ist dir denn passiert?\" fragte der Esel. Der Hund und die Katze erzählten ihre Geschichte, und die Katze ging auch mit nach Bremen.",
				Translation: "一头驴子忠诚地效力多年，但如今它老了，干不动活了。主人不想再留它，于是驴子离家出走，前往不来梅。它想，到了那里可以当个城市音乐家。\n\n" +
					"它走了一阵，在路上遇见一只猎狗，趴在地上喘着粗气，像是跑累了。“兄弟，你怎么喘成这样？”驴子问。“唉，”猎狗说，“因为我老了，可还得天天打猎，主人用不上我了，就把我赶了出来。”\n\n" +
					"“知道吗？”驴子说，“我要去不来梅当城市音乐家。跟我一起走吧，也去乐队谋个差事。我弹鲁特琴，你敲鼓。”猎狗同意了，它们便一起上路。\n\n" +
					"不久之后，一只猫蹲在路边，摆着一副苦大仇深的脸色。“你怎么了？”驴子问。听猎狗和它讲完各自的遭遇后，猫也一同前往不来梅。",
			},
			{
				BookID: bremen.ID, Sort: 2, Status: 1,
				Title: "Kapitel 2 · Das Haus der Räuber 强盗的房子",
				Content: "Danach kamen die drei Tiere an einem Hof vorbei, da saß ein Hahn auf dem Tor und schrie aus Leibeskräften. " +
					"\"Dein Schreien geht einem durch Mark und Bein\", sagte der Esel, \"was ist denn los?\" " +
					"Der Hahn erzählte, dass er den guten Morgen verkündige, weil die Räuber im Haus friedlich schliefen.\n\n" +
					"Die vier Tiere beschlossen zusammenzubleiben. In dem Haus war es warm und es gab viel zu essen, sagten sie. " +
					"Am Abend kamen sie zu dem Haus der Räuber. Es war hell darin, und die Räuber saßen am Tisch und aßen und tranken.\n\n" +
					"Der Esel stellte sich mit den Vorderfüßen auf das Fenster, der Hund sprang auf des Esels Rücken, " +
					"die Katze kroch auf den Hund, und endlich flog der Hahn hinauf und setzte sich auf den Kopf der Katze. " +
					"Dann fingen sie alle zusammen an zu musizieren. Die Räuber fuhren erschrocken hoch und rannten in den Wald hinaus.\n\n" +
					"Die vier Musikanten setzten sich an den Tisch, aßen die Reste mit Freude und fielen danach müde in Schlaf.",
				Translation: "随后，三个动物路过一个院子，一只公鸡蹲在大门上拼命啼叫。“你这一嗓子听得人骨头都酥了，”驴子说，“你叫什么呢？”公鸡解释说，它在报晓，因为强盗们正在屋里安睡。\n\n" +
					"四个动物决定结伴同行。它们说，那房子里暖和，还有好多吃的。傍晚，它们来到强盗的房子前。屋里亮着灯，强盗们坐在桌边又吃又喝。\n\n" +
					"驴子把前脚搭上窗台，狗跳到驴背上，猫爬到狗身上，公鸡最后飞上去站在猫头上了。然后它们齐声开始“演奏”。强盗们吓得跳起来，逃进了森林。\n\n" +
					"四位音乐家坐到桌边，高高兴兴地吃掉了剩菜，然后疲惫地进入了梦乡。",
			},
		})
	}

	// 小说专栏示例（原创短篇，可直接删除后导入你自己的小说）
	var novelCount int64
	db.Model(&Book{}).Where("type = 1").Count(&novelCount)
	if novelCount == 0 {
		novel := Book{
			Title:    "小站的灯",
			Subtitle: "The Lamp at the Small Station",
			Author:   "示例小说（可删除）",
			Language: "zh",
			Description: "一个关于守站人与夜车灯火的短篇故事。导入自己的小说请到后台" +
				"「书籍管理」新建书籍时选择类型为「小说专栏」，再用批量导入章节。",
			Type:   1,
			Status: 1,
			Sort:   1,
		}
		db.Create(&novel)

		db.Create(&[]Chapter{
			{
				BookID: novel.ID, Sort: 1, Status: 1,
				Title: "第一章 雪夜",
				Content: "雪是从傍晚开始下的。\n\n小站上只有一盏灯还亮着，昏黄的光落在铁轨上，像谁不小心打翻了一小罐蜂蜜。老周把炉子里的煤块拨了拨，火星子噼啪跳了两下，又安静下去。\n\n" +
					"这是他在这个站的第三十一个冬天。白天还有两趟慢车会停，到了夜里，整个站台上就只剩下他、风，和那条趴在值班室门口的老黄狗。\n\n" +
					"九点四十，末班车会经过，不停。但老周还是会提着灯走到站台上去，把灯举得高高的。他说不清为什么，只是觉得，黑漆漆的野地里，总该有一盏灯是亮着的。\n\n" +
					"火车呼啸而过的时候，他看见第三节车厢的窗户里，有个孩子把脸贴在玻璃上，朝他用力挥了挥手。\n\n老周也挥了挥手。雪落在他的帽檐上，落了薄薄一层。",
			},
			{
				BookID: novel.ID, Sort: 2, Status: 1,
				Title: "第二章 旅客",
				Content: "那年冬天特别长。\n\n二月里的一个深夜，一列本不该停的货车在小站临时停了下来，说是前方线路出了故障。押车的年轻人跳下车来跺脚取暖，看见了值班室里的灯，便敲门进来讨口热水。\n\n" +
					"老周给他倒了水，又从炉边烤着的几个红薯里拿了一个递过去。年轻人愣了愣，笑了，说这一路上，就这一口热的最实在。\n\n" +
					"两个人有一搭没一搭地聊着。年轻人说他跑这条线三年了，每次夜里经过这个小站，都会看见站台上有一盏灯，和一个举着灯的人影。\n\n" +
					"「我一直以为那是个信号，」他咬着红薯说，「后来才知道，什么信号也不是。」\n\n老周往炉子里添了块煤，慢悠悠地说：「灯不是给信号亮的，是给人亮的。夜里赶路的人，看见一点亮，心里就不慌。」\n\n货车修好后继续往前开。年轻人从车窗里探出头，朝站台敬了个不标准的礼。",
			},
			{
				BookID: novel.ID, Sort: 3, Status: 1,
				Title: "第三章 交班",
				Content: "春天来的时候，老周退休了。\n\n新来的值守员是个二十出头的年轻人，报到那天，老周带他把站里站外走了一遍，最后把那盏用了三十年的旧提灯交到他手里。\n\n" +
					"「夜里九点四十，末班车过站，不用停，」老周说，「但灯还是要举的。」\n\n年轻人问为什么。\n\n" +
					"老周想了很久，说：「也说不上为什么。就是黑夜里跑着的那些人，说不定有谁，正需要看见一点亮光。」\n\n" +
					"那天夜里，新值守员提着灯走上站台。火车呼啸而过，有一节车厢的玻璃后面，映出一小片暖黄的光，一闪，就消失在春天的夜色里了。",
			},
		})
	}

	// 动漫排行榜种子（示例数据，后台「动漫管理」可增删改）
	var animeCount int64
	db.Model(&Anime{}).Count(&animeCount)
	if animeCount == 0 {
		db.Create(&[]Anime{
			{Title: "葬送的芙莉莲", Region: "日本", Episodes: "更新至第28集", Rank: 1, Status: 1,
				Description: "勇者一行冒险结束后的后日谈，魔法与时光的温柔物语。"},
			{Title: "孤独摇滚", Region: "日本", Episodes: "全12集", Rank: 2, Status: 1,
				Description: "社恐少女的摇滚乐队日常，治愈又爆笑。"},
			{Title: "夏日重现", Region: "日本", Episodes: "全25集", Rank: 3, Status: 1,
				Description: "时间循环 × 小岛悬疑，夏天必备神作。"},
			{Title: "灵能百分百", Region: "日本", Episodes: "全37集", Rank: 4, Status: 1,
				Description: "超能力少年的成长物语，作画顶级。"},
			{Title: "紫罗兰永恒花园", Region: "日本", Episodes: "全13集", Rank: 5, Status: 1,
				Description: "代笔少女学会「爱」的过程，京都动画画面巅峰。"},
		})
	}

	// 艺术鉴赏种子（示例占位作品，后台「艺术鉴赏管理」可增删改）
	var artCount int64
	db.Model(&Artwork{}).Count(&artCount)
	if artCount == 0 {
		db.Create(&[]Artwork{
			{Title: "示例作品 · 星夜练习", Description: "这是一条示例简介：上传图片并填写标题与简介，前台画廊会自动展示。", Author: "小天", Sort: 1, Status: 1},
		})
	}

	// AI 工具导航种子（AI编程大类先行，后台可增删改）
	var aiToolCount int64
	db.Model(&AiTool{}).Count(&aiToolCount)
	if aiToolCount == 0 {
		db.Create(&[]AiTool{
			{Name: "ChatGPT", URL: "https://chatgpt.com", Section: "AI编程", Category: "对话助手",
				Description: "OpenAI 旗舰对话模型，编程问答、代码生成皆可。", Tags: "付费,免费额度", Sort: 1, Status: 1},
			{Name: "Claude", URL: "https://claude.ai", Section: "AI编程", Category: "对话助手",
				Description: "Anthropic 出品，长上下文与代码理解出色。", Tags: "付费,免费额度", Sort: 2, Status: 1},
			{Name: "GitHub Copilot", URL: "https://github.com/features/copilot", Section: "AI编程", Category: "代码补全",
				Description: "最流行的 AI 结对编程助手，IDE 内实时代码补全。", Tags: "付费,学生免费", Sort: 3, Status: 1},
			{Name: "Cursor", URL: "https://cursor.com", Section: "AI编程", Category: "AI IDE",
				Description: "AI 原生代码编辑器，整个项目上下文理解。", Tags: "付费,免费额度", Sort: 4, Status: 1},
			{Name: "通义灵码", URL: "https://tongyi.aliyun.com/lingma", Section: "AI编程", Category: "代码补全",
				Description: "阿里出品，中文注释生成友好，个人版免费。", Tags: "免费", Sort: 5, Status: 1},
			{Name: "v0", URL: "https://v0.dev", Section: "AI编程", Category: "前端生成",
				Description: "Vercel 出品，描述需求直接生成 React/Tailwind 界面。", Tags: "付费,免费额度", Sort: 6, Status: 1},
		})
		db.Create(&[]AiPrompt{
			{Title: "代码审查助手", Section: "AI编程", Category: "编程",
				Content: "你是一位资深代码审查员。请审查我提供的代码，从：1) 潜在 bug 2) 性能问题 3) 可读性 4) 安全性 四个方面给出具体改进建议，用中文回答，代码保持原语言。",
				Description: "让 AI 按四个维度系统性审查代码", Sort: 1, Status: 1},
			{Title: "报错诊断", Section: "AI编程", Category: "编程",
				Content: "我遇到如下报错，请分析可能的原因并给出排查步骤，按可能性从高到低排列：\n\n[粘贴报错信息]",
				Description: "把报错扔给 AI 快速定位", Sort: 2, Status: 1},
		})
	}

	// 游戏板块种子（电竞分类先行，后台「游戏管理」可增删改）
	var gameCount int64
	db.Model(&Game{}).Count(&gameCount)
	if gameCount == 0 {
		db.Create(&[]Game{
			{Title: "英雄联盟", Category: "电竞", Platform: "PC", Tags: "MOBA,免费",
				Description: "全球最流行的 MOBA 电竞项目，S 赛是全球电竞狂欢。", Hot: 980000, Sort: 1, Status: 1},
			{Title: "CS2", Category: "电竞", Platform: "PC", Tags: "FPS,免费",
				Description: "经典竞技射击的续作，Major 赛事体系成熟。", Hot: 870000, Sort: 2, Status: 1},
			{Title: "王者荣耀", Category: "电竞", Platform: "手机", Tags: "MOBA,免费",
				Description: "国民手游电竞，KPL 职业联赛热度极高。", Hot: 950000, Sort: 3, Status: 1},
			{Title: "无畏契约", Category: "电竞", Platform: "PC", Tags: "FPS,免费",
				Description: " Riot 出品的战术射击，技能+枪法组合竞技。", Hot: 760000, Sort: 4, Status: 1},
			{Title: "原神", Category: "开放世界", Platform: "全平台", Tags: "RPG,免费",
				Description: "米哈游开放世界标杆，探索与剧情兼备。", Hot: 890000, Sort: 5, Status: 1},
			{Title: "艾尔登法环", Category: "动作冒险", Platform: "主机,PC", Tags: "魂系,买断",
				Description: "宫崎英高的开放魂系神作，DLC 黄金树幽影同样精彩。", Hot: 720000, Sort: 6, Status: 1},
			{Title: "塞尔达传说：王国之泪", Category: "开放世界", Platform: "Switch", Tags: "任天堂,买断",
				Description: "创造力天花板，究极手玩法自由度无上限。", Hot: 810000, Sort: 7, Status: 1},
			{Title: "黑神话：悟空", Category: "动作冒险", Platform: "PC,PS5", Tags: "国产,买断",
				Description: "国产 3A 里程碑，西游题材动作 RPG。", Hot: 930000, Sort: 8, Status: 1},
		})
	}
}