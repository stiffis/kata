package generator

import (
	"math/rand"
	"os"
	"strings"
	"time"
)

type LessonType int

const (
	TypeBigrams LessonType = iota
	TypeWords
	TypeSymbols
	TypeCode
	TypeFile
	TypeWeaknesses
)

type Language string

const (
	LangGo         Language = "go"
	LangEnglish    Language = "english"
	LangSpanish    Language = "spanish"
	LangFrench     Language = "french"
	LangGerman     Language = "german"
	LangPython     Language = "python"
	LangCpp        Language = "cpp"
	LangJavascript Language = "javascript"
	LangRust       Language = "rust"
)

type Generator struct {
	rand     *rand.Rand
	Language Language
}

type WeakKey struct {
	Key       string
	ErrorRate float64
}

func New() *Generator {
	return &Generator{
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
		Language: LangGo, // Default
	}
}

func (g *Generator) SetLanguage(lang Language) {
	g.Language = lang
}

// Data repositories
var commonBigrams = []string{
	"th", "he", "in", "er", "an", "re", "on", "at", "en", "nd", "ti", "es", "or", "te", "of", "ed", "is", "it", "al", "ar", "st", "to", "nt", "ng", "se", "ha", "as", "ou", "io", "le", "ve", "me", "ea", "hi", "ne", "de", "ra", "co",
}

var spanishWords = []string{
	"que", "de", "no", "a", "la", "el", "es", "y", "en", "lo", "un", "por", "qué", "me", "una", "te", "los", "se", "con", "para", "mi", "está", "si", "bien", "pero", "yo", "eso", "las", "sí", "su", "tu", "aquí", "del", "al", "como", "le", "más", "esto", "ya", "todo", "esta", "vamos", "muy", "hay", "ahora", "algo", "estoy", "tengo", "nos", "tú", "nada", "cuando", "ha", "este", "sé", "estás", "así", "puedo", "cómo", "quiero", "sólo", "soy", "tiene", "gracias", "o", "él", "bueno", "fue", "ser", "hacer", "son", "todos", "era", "eres", "vez", "tienes", "creo", "ella", "he", "ese", "voy", "puede", "sabes", "hola", "sus", "porque", "dios", "quién", "nunca", "dónde", "quieres", "casa", "favor", "esa", "dos", "tan", "señor", "tiempo", "verdad", "estaba", "mejor", "están", "va", "hombre", "usted", "mucho", "hace", "entonces", "siento", "tenemos", "picos", "donde",
	"vida", "ver", "alguien", "siempre", "hasta", "sin", "mismo", "han", "trabajo", "noche", "mundo", "parte", "padre", "tres", "gente", "decir", "hijo", "ni", "realmente", "amor", "lugar", "dinero", "buena", "amigo", "gran", "nuevo", "cosa", "tipo", "mira", "después", "mañana", "nuestro", "hizo", "madre", "nadie", "dicen", "mal", "hoy", "nosotros", "mía", "otras", "fuego", "cabeza", "cualquier", "manera", "día", "fin", "sola", "seguro", "historia", "luego", "cuenta", "quizás", "mujer", "claro", "año", "menos", "familia", "importa", "contar", "problema", "razón", "agua", "papá", "poder", "cinco", "fuerza", "ciudad", "guerra", "mano", "nombre", "chica", "cuerpo", "lado", "sido", "hablar", "hombres", "tarde", "estas", "único", "muerte", "chico", "minutos", "frente", "pueblo", "sueño", "quien", "fuerte", "ojos", "país", "semana", "mes", "cerca", "libro", "voz", "escuela", "juego", "puerta", "camino", "aire", "coche", "tierra", "paso", "sol", "luz", "alma", "alto", "rojo", "causa", "papel", "joven", "punto", "color", "cara", "final", "grande", "vivir", "pasar", "comer", "salir", "entrar", "tomar", "dormir", "morir", "leer", "cambiar", "sentir",
}

var frenchWords = []string{
	"je", "de", "est", "pas", "le", "vous", "la", "tu", "que", "un", "il", "et", "à", "a", "ne", "les", "en", "ce", "ça", "une", "ai", "pour", "on", "moi", "des", "mais", "bien", "du", "nous", "y", "me", "dans", "c'est", "elle", "si", "tout", "plus", "non", "mon", "suis", "te", "au", "avec", "va", "qui", "oui", "fait", "ils", "faire", "ma", "comme", "être", "sur", "quoi", "toi", "ici", "rien", "dit", "lui", "votre", "sais", "bon", "là", "pourquoi", "quand", "as", "étais", "peux", "son", "aussi", "avez", "par", "voir", "merci", "ont", "jamais", "où", "aller", "sont", "cette", "dire", "se", "veux", "tous", "m'a", "peut", "comment", "très", "même", "ta", "chose",
	"vie", "deux", "temps", "peu", "homme", "monde", "encore", "vrai", "notre", "mal", "sa", "faut", "fois", "sans", "grand", "toujours", "elle", "seul", "autre", "femme", "après", "trouver", "jour", "ans", "père", "mère", "fils", "fille", "gens", "trop", "mes", "besoin", "accord", "ses", "mieux", "tes", "voir", "mort", "nuit", "main", "place", "beau", "maison", "nom", "travail", "trois", "petit", "choses", "personne", "heure", "avoir", "juste", "crois", "ton", "toute", "alors", "bonne", "penser", "prendre", "venir", "regarder", "demander", "laisser", "partir", "mettre", "rester", "aimer", "pouvoir", "vouloir", "devoir", "savoir", "donner", "comprendre", "connaître", "entendre", "parler", "jouer", "passer", "travailler", "manger", "dormir", "boire", "lire", "écrire", "acheter", "payer", "aider", "attendre", "finir", "perdre", "gagner", "sentir", "vivre", "mourir", "ouvrir", "fermer", "chercher", "trouver", "marcher", "courir", "sauter", "danser", "chanter", "rire", "pleurer", "sourire", "apprendre", "enseigner", "étudier", "oublier", "rappeler",
}

var germanWords = []string{
	"ich", "ist", "nicht", "das", "du", "es", "sie", "und", "der", "wir", "was", "ein", "zu", "er", "in", "mir", "mit", "ja", "wie", "den", "auf", "mich", "dass", "aber", "eine", "so", "hat", "hier", "habe", "für", "sind", "wenn", "nein", "von", "dich", "war", "haben", "an", "einen", "uns", "da", "hab", "bin", "noch", "dir", "man", "nur", "sich", "ihr", "kann", "dem", "muss", "schon", "wer", "sein", "jetzt", "dann", "die", "immer", "mal", "wird", "als", "nichts", "alles", "doch", "gut", "nach", "aus", "um", "mein", "also", "ihm", "weiß", "wieder", "tun", "will", "keine", "geht", "mehr", "warum", "gesagt", "morgen", "bitte", "vor", "bei", "alle", "einer", "los", "vielleicht", "wäre", "wo", "hast", "also",
	"Zeit", "Jahr", "Mensch", "Leben", "Kind", "Tag", "Hand", "Auge", "Land", "Wort", "Gott", "Welt", "Stadt", "Frau", "Herr", "Mann", "Haus", "Ding", "Ende", "Fall", "Weg", "Teil", "Woche", "Name", "Vater", "Mutter", "Sohn", "Tochter", "Freund", "Arbeit", "Geld", "Leute", "Nacht", "Wasser", "Seite", "Schule", "Problem", "Angst", "Recht", "Stunde", "Kopf", "Bild", "Buch", "Tisch", "Stuhl", "Raum", "Platz", "Musik", "Kunst", "Spiel", "Film", "Brief", "Farbe", "Licht", "Luft", "Boden", "Feuer", "Holz", "Stein", "Glas", "Gold", "Silber", "Eisen", "Stahl", "Kupfer", "Blei", "Zinn", "Zink", "Aluminium", "Bronze", "Messing", "Nickel", "Chrom", "Titan", "Platin", "Uran", "Plutonium", "Helium", "Neon", "Argon", "Krypton", "Xenon", "Radon", "Wasserstoff", "Sauerstoff", "Stickstoff", "Kohlenstoff", "Schwefel", "Phosphor", "Chlor", "Jod", "Fluor", "Brom", "gehen", "kommen", "sehen", "hören", "sagen", "machen", "geben", "nehmen", "lassen", "stehen", "liegen", "sitzen", "bleiben", "glauben", "denken", "wissen", "kennen", "verstehen", "lernen", "arbeiten", "spielen", "singen", "laufen", "fahren", "fliegen", "schwimmen", "essen", "trinken", "schlafen", "lesen", "schreiben", "malen", "bauen", "kaufen",
}

var englishWords = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "i", "it", "for", "not", "on", "with", "he", "as", "you", "do", "at", "this", "but", "his", "by", "from", "they", "we", "say", "her", "she", "or", "an", "will", "my", "one", "all", "would", "there", "their", "what", "so", "up", "out", "if", "about", "who", "get", "which", "go", "me", "when", "make", "can", "like", "time", "no", "just", "him", "know", "take", "people", "into", "year", "your", "good", "some", "could", "them", "see", "other", "than", "then", "now", "look", "only", "come", "its", "over", "think", "also", "back", "after", "use", "two", "how", "our", "work", "first", "well", "way", "even", "new", "want", "because", "any", "these", "give", "day", "most", "us",
	"life", "child", "world", "school", "state", "family", "student", "group", "country", "problem", "hand", "part", "place", "case", "week", "company", "system", "program", "question", "work", "government", "number", "night", "point", "home", "water", "room", "mother", "area", "money", "story", "fact", "month", "lot", "right", "study", "book", "eye", "job", "word", "business", "issue", "side", "kind", "head", "house", "service", "friend", "father", "power", "hour", "game", "line", "end", "member", "law", "car", "city", "community", "name", "president", "team", "minute", "idea", "kid", "body", "information", "back", "parent", "face", "others", "level", "office", "door", "health", "person", "art", "war", "history", "party", "result", "change", "morning", "reason", "research", "girl", "guy", "moment", "air", "teacher", "force", "education", "foot", "boy", "age", "policy", "everything", "process", "music", "market", "sense", "nation", "plan", "college", "interest", "death", "experience", "effect", "use", "class", "control", "care", "field", "development", "role", "effort", "rate", "heart", "drug", "show", "leader", "light", "voice", "wife", "police", "mind", "price", "report", "decision", "son", "view", "relationship", "town", "road", "arm", "difference", "value", "building", "action", "model", "season", "society", "tax", "director", "position", "player", "agree", "allow", "answer", "ask", "become", "begin", "believe", "borrow", "break", "bring", "build", "buy", "call", "carry", "catch", "choose", "close", "cut", "decide", "die", "draw", "drink", "drive", "eat", "explain", "fall", "feel", "fill", "find", "finish", "fly", "follow", "forget", "grow", "happen", "hear", "help", "hold", "hope", "keep", "kill", "learn", "leave", "let", "lie", "listen", "live", "lose", "love", "mean", "meet", "move", "need", "open", "pay", "play", "promise", "put", "reach", "read", "remember", "run", "sell", "send", "set", "sit", "speak", "spend", "stand", "start", "stop", "suggest", "talk", "teach", "tell", "think", "travel", "try", "turn", "understand", "wait", "walk", "watch", "win", "wonder", "work", "write", "worry",
}

var pythonKeywords = []string{
	"def", "class", "import", "return", "if", "else", "elif", "for", "while", "try", "except", "finally", "with", "as", "lambda", "yield", "True", "False", "None", "self", "print", "len", "range", "list", "dict", "set", "str", "int", "float", "__init__",
	"super", "is", "in", "not", "and", "or", "break", "continue", "pass", "raise", "global", "nonlocal", "del", "from", "assert", "async", "await", "open", "read", "write", "close", "append", "extend", "pop", "remove", "sort", "reverse", "join", "split", "strip", "replace", "find", "count", "lower", "upper", "title", "keys", "values", "items", "get", "update", "clear", "copy", "map", "filter", "reduce", "zip", "enumerate", "sorted", "reversed", "min", "max", "sum", "abs", "round", "pow", "divmod", "all", "any", "isinstance", "issubclass", "hasattr", "getattr", "setattr", "delattr", "callable", "type", "id", "hash", "dir", "help", "vars", "locals", "globals", "input", "eval", "exec", "repr", "str", "bytes", "bytearray", "memoryview", "bool", "complex", "object", "property", "staticmethod", "classmethod", "importlib", "sys", "os", "math", "random", "datetime", "time", "json", "re", "collections", "itertools", "functools", "pathlib", "typing", "argparse", "logging", "subprocess", "threading", "multiprocessing", "asyncio", "socket", "requests", "numpy", "pandas", "matplotlib", "django", "flask", "sqlalchemy", "pytest", "unittest",
}

var cppKeywords = []string{
	"int", "char", "float", "double", "bool", "void", "wchar_t", "class", "struct", "union", "enum", "template", "typename", "namespace", "using", "public", "private", "protected", "virtual", "friend", "this", "new", "delete", "operator", "true", "false", "nullptr", "if", "else", "for", "while", "do", "switch", "case", "default", "break", "continue", "return", "try", "catch", "throw", "const", "static", "volatile", "mutable", "extern", "register", "auto", "sizeof", "typedef", "std::cout", "std::cin", "std::vector", "std::string",
	"std::map", "std::set", "std::list", "std::array", "std::pair", "std::tuple", "std::unique_ptr", "std::shared_ptr", "std::weak_ptr", "std::make_unique", "std::make_shared", "std::move", "std::forward", "std::function", "std::bind", "std::thread", "std::mutex", "std::lock_guard", "std::atomic", "std::condition_variable", "std::future", "std::promise", "std::chrono", "std::filesystem", "std::optional", "std::variant", "std::any", "std::string_view", "constexpr", "consteval", "constinit", "decltype", "noexcept", "override", "final", "explicit", "inline", "static_assert", "alignas", "alignof", "typeid", "dynamic_cast", "static_cast", "reinterpret_cast", "const_cast", "goto", "asm", "concept", "requires", "co_await", "co_yield", "co_return", "module", "import", "export", "#include", "#define", "#ifdef", "#ifndef", "#endif", "#pragma", "main", "argc", "argv", "std::endl", "push_back", "emplace_back", "pop_back", "begin", "end", "cbegin", "cend", "rbegin", "rend", "size", "empty", "clear", "resize", "reserve", "find", "count", "insert", "erase", "swap", "sort",
}

var jsKeywords = []string{
	"var", "let", "const", "function", "return", "if", "else", "for", "while", "do", "switch", "case", "default", "break", "continue", "try", "catch", "finally", "throw", "new", "this", "super", "class", "extends", "import", "export", "default", "async", "await", "yield", "void", "typeof", "instanceof", "in", "of", "delete", "true", "false", "null", "undefined", "NaN", "Infinity", "console.log", "document", "window", "Promise", "JSON",
	"map", "filter", "reduce", "forEach", "find", "findIndex", "includes", "indexOf", "join", "split", "slice", "splice", "push", "pop", "shift", "unshift", "concat", "sort", "reverse", "some", "every", "fill", "from", "isArray", "keys", "values", "entries", "assign", "create", "defineProperty", "freeze", "seal", "toString", "parseInt", "parseFloat", "setTimeout", "setInterval", "clearTimeout", "clearInterval", "addEventListener", "removeEventListener", "getElementById", "querySelector", "querySelectorAll", "createElement", "appendChild", "removeChild", "classList", "setAttribute", "getAttribute", "innerHTML", "innerText", "textContent", "fetch", "then", "catch", "resolve", "reject", "all", "race", "any", "allSettled", "Math", "Date", "RegExp", "Error", "Map", "Set", "WeakMap", "WeakSet", "Symbol", "BigInt", "arrow", "callback", "closure", "prototype", "constructor", "module", "require", "exports",
}

var rustKeywords = []string{
	"fn", "let", "mut", "const", "static", "if", "else", "match", "loop", "while", "for", "in", "return", "break", "continue", "struct", "enum", "trait", "impl", "use", "mod", "pub", "crate", "super", "self", "Self", "type", "where", "unsafe", "extern", "async", "await", "move", "ref", "box", "true", "false", "None", "Some", "Ok", "Err", "String", "Vec", "Option", "Result", "println!", "format!", "vec!", "pub",
	"panic!", "assert!", "assert_eq!", "dbg!", "todo!", "unreachable!", "macro_rules!", "derive", "Debug", "Clone", "Copy", "PartialEq", "Eq", "PartialOrd", "Ord", "Hash", "Default", "Display", "Error", "From", "Into", "AsRef", "AsMut", "Drop", "Send", "Sync", "Box", "Rc", "Arc", "Cell", "RefCell", "Mutex", "RwLock", "HashMap", "HashSet", "BTreeMap", "BTreeSet", "LinkedList", "BinaryHeap", "Duration", "Instant", "Thread", "File", "Path", "PathBuf", "io", "fs", "env", "args", "Result", "Option", "u8", "u16", "u32", "u64", "u128", "usize", "i8", "i16", "i32", "i64", "i128", "isize", "f32", "f64", "bool", "char", "str", "slice", "iter", "collect", "map", "filter", "fold", "zip", "enumerate", "unwrap", "expect", "unwrap_or", "unwrap_or_else", "and_then", "ok_or", "as_str", "as_bytes", "to_string", "to_owned", "push", "pop", "insert", "remove", "contains", "len", "is_empty",
}

var goKeywords = []string{
	"func", "package", "import", "return", "if", "else", "for", "range",
	"struct", "interface", "type", "var", "const", "go", "defer", "map",
	"chan", "select", "case", "switch", "break", "continue", "fallthrough",
	"goto", "true", "false", "nil", "error", "string", "int", "bool",
	"byte", "rune", "float32", "float64", "complex64", "complex128", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "int8", "int16", "int32", "int64", "make", "new", "len", "cap", "append", "copy", "close", "delete", "complex", "real", "imag", "panic", "recover", "print", "println", "fmt", "os", "io", "bufio", "net", "http", "json", "xml", "sql", "time", "log", "sync", "atomic", "context", "errors", "flag", "path", "filepath", "sort", "strings", "strconv", "unicode", "math", "rand", "crypto", "sha256", "aes", "bytes", "reflect", "unsafe", "syscall", "exec", "signal", "user", "template", "html", "image", "png", "jpeg", "gif", "testing", "pprof", "runtime", "debug", "main", "init", "Error", "String", "Write", "Read", "Close", "ServeHTTP", "Handler", "Context", "Done", "Err", "Value", "Type", "Kind", "Field", "Method", "Add", "Done", "Wait", "Lock", "Unlock", "RLock", "RUnlock", "Once", "Pool", "Map", "Cond", "WaitGroup", "Mutex", "RWMutex",
}

var goSnippets = []string{
	"func main() {\n\tfmt.Println(\"Hello\")\n}",
	"if err != nil {\n\treturn err\n}",
	"for i := 0; i < len(items); i++ {\n\tfmt.Println(items[i])\n}",
	"for _, item := range items {\n\tprocess(item)\n}",
	"type User struct {\n\tID   int\n\tName string\n}",
	"func (u *User) GetName() string {\n\treturn u.Name\n}",
	"ch := make(chan int, 10)",
	"defer file.Close()",
	"result, err := doSomething()",
	"switch value {\ncase 1:\n\treturn \"one\"\ndefault:\n\treturn \"other\"\n}",
	"http.HandleFunc(\"/\", func(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprintf(w, \"Hi\")\n})",
	"ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)\ndefer cancel()",
	"var wg sync.WaitGroup\nwg.Add(1)\ngo func() {\n\tdefer wg.Done()\n\twork()\n}()\nwg.Wait()",
	"if _, err := os.Stat(filename); os.IsNotExist(err) {\n\treturn fmt.Errorf(\"file not found\")\n}",
	"type Service interface {\n\tDo(ctx context.Context) error\n}",
	"select {\ncase <-ctx.Done():\n\treturn ctx.Err()\ncase val := <-ch:\n\tfmt.Println(val)\n}",
	"b, err := json.Marshal(data)\nif err != nil {\n\tlog.Fatal(err)\n}",
	"func init() {\n\tflag.StringVar(&addr, \"addr\", \":8080\", \"address\")\n}",
	"mu.Lock()\ndefer mu.Unlock()\ncount++",
	"file, err := os.Open(\"file.txt\")\nif err != nil {\n\treturn err\n}",
}

var pythonSnippets = []string{
	"def main():\n    print(\"Hello\")",
	"if __name__ == \"__main__\":\n    main()",
	"for i in range(10):\n    print(i)",
	"class User:\n    def __init__(self, name):\n        self.name = name",
	"try:\n    value = int(x)\nexcept ValueError:\n    value = 0",
	"with open('file.txt') as f:\n    content = f.read()",
	"items = [x * 2 for x in numbers]",
	"def get_name(self):\n    return self.name",
	"import json\nwith open('data.json', 'w') as f:\n    json.dump(data, f)",
	"async def fetch_data():\n    async with aiohttp.ClientSession() as session:\n        pass",
	"def decorator(func):\n    def wrapper(*args, **kwargs):\n        return func(*args, **kwargs)\n    return wrapper",
	"@decorator\ndef my_func():\n    pass",
	"lambda x: x * 2",
	"users = {u.id: u for u in user_list}",
	"if x is None:\n    x = []",
	"for index, item in enumerate(items):\n    print(f\"{index}: {item}\")",
	"import os\nfiles = [f for f in os.listdir('.') if f.endswith('.py')]",
	"class Meta(type):\n    def __new__(cls, name, bases, dct):\n        return super().__new__(cls, name, bases, dct)",
	"from typing import List, Optional\n\ndef process(items: List[int]) -> Optional[int]:\n    return items[0] if items else None",
	"try:\n    pass\nfinally:\n    cleanup()",
}

var cppSnippets = []string{
	"#include <iostream>\n\nint main() {\n    std::cout << \"Hello World\";\n    return 0;\n}",
	"template <typename T>\nT add(T a, T b) {\n    return a + b;\n}",
	"class MyClass {\npublic:\n    void myMethod() {\n        // Code here\n    }\n};",
	"for (int i = 0; i < 10; ++i) {\n    vec.push_back(i);\n}",
	"if (ptr != nullptr) {\n    delete ptr;\n    ptr = nullptr;\n}",
	"std::vector<int> numbers = {1, 2, 3, 4, 5};",
	"const std::string& getName() const {\n    return name;\n}",
	"for (const auto& item : items) {\n    std::cout << item << std::endl;\n}",
	"auto lambda = [](int x) { return x * 2; };",
	"std::sort(vec.begin(), vec.end(), [](int a, int b) {\n    return a > b;\n});",
	"class Derived : public Base {\n    void overrideMe() override;\n};",
	"std::unique_ptr<int> ptr = std::make_unique<int>(10);",
	"std::lock_guard<std::mutex> lock(mu);",
	"using StringMap = std::map<std::string, std::string>;",
	"template <typename... Args>\nvoid print(Args... args) {\n    (std::cout << ... << args) << '\\n';\n}",
	"if (auto it = map.find(key); it != map.end()) {\n    // Found\n}",
	"struct Point {\n    int x, y;\n    auto operator<=>(const Point&) const = default;\n};",
	"constexpr int factorial(int n) {\n    return n <= 1 ? 1 : n * factorial(n - 1);\n}",
	"try {\n    throw std::runtime_error(\"Error\");\n} catch (const std::exception& e) {\n    std::cerr << e.what();\n}",
}

var jsSnippets = []string{
	"console.log(\"Hello World\");",
	"const add = (a, b) => a + b;",
	"document.getElementById(\"app\").innerHTML = \"Hi\";",
	"async function fetchData() {\n  const res = await fetch(url);\n  return res.json();\n}",
	"class User extends Person {\n  constructor(name) {\n    super(name);\n  }\n}",
	"items.map(item => item * 2).filter(i => i > 10);",
	"if (obj && obj.prop) {\n  return obj.prop;\n}",
	"import React, { useState } from 'react';",
	"const [count, setCount] = useState(0);",
	"useEffect(() => {\n  document.title = `Count: ${count}`;\n}, [count]);",
	"const { name, age, ...rest } = user;",
	"return new Promise((resolve, reject) => {\n  setTimeout(resolve, 1000);\n});",
	"try {\n  await doWork();\n} catch (e) {\n  console.error(e);\n}",
	"export const myFunc = () => true;",
	"const merged = { ...obj1, ...obj2 };",
	"document.querySelector('.btn').addEventListener('click', e => {\n  e.preventDefault();\n});",
	"const str = `User: ${user.name}`;",
	"if (Array.isArray(data)) {\n  data.forEach(d => console.log(d));\n}",
	"module.exports = { foo, bar };",
	"const regex = /^[a-z]+$/i;",
}

var rustSnippets = []string{
	"fn main() {\n    println!(\"Hello, world!\");\n}",
	"let mut x = 5;\nprintln!(\"The value of x is: {}\", x);",
	"match number {\n    1 => println!(\"One!\"),\n    _ => println!(\"Ain't one\"),\n}",
	"pub struct Point {\n    x: i32,\n    y: i32,\n}",
	"impl Rectangle {\n    fn area(&self) -> u32 {\n        self.width * self.height\n    }\n}",
	"let v: Vec<i32> = Vec::new();",
	"if let Some(i) = option {\n    println!(\"Matched {:?}\", i);\n}",
	"pub fn add(left: usize, right: usize) -> usize {\n    left + right\n}",
	"use std::io::{self, Write};\nlet stdout = io::stdout();",
	"#[derive(Debug, Clone, PartialEq)]\nstruct User {\n    name: String,\n}",
	"let numbers: Vec<i32> = (0..10).map(|x| x * 2).collect();",
	"fn longer<'a>(s1: &'a str, s2: &'a str) -> &'a str {\n    if s1.len() > s2.len() { s1 } else { s2 }\n}",
	"pub trait Summary {\n    fn summarize(&self) -> String;\n}",
	"match result {\n    Ok(val) => println!(\"Success: {}\", val),\n    Err(e) => println!(\"Error: {}\", e),\n}",
	"let handle = thread::spawn(|| {\n    for i in 1..10 {\n        println!(\"hi number {} from the spawned thread!\", i);\n    }\n});",
	"let mut map = HashMap::new();\nmap.insert(\"Blue\", 10);",
	"async fn learn_and_sing() -> Song {\n    let song = learn_song().await;\n    sing_song(song).await\n}",
	"macro_rules! say_hello {\n    () => {\n        println!(\"Hello!\");\n    };\n}",
	"let s = String::from(\"hello\");\nlet len = calculate_length(&s);",
	"enum Message {\n    Quit,\n    Move { x: i32, y: i32 },\n    Write(String),\n    ChangeColor(i32, i32, i32),\n}",
}

var goSymbols = []string{
	"{}", "[]", "()", "!=", "==", "<=", ">=", "&&", "||", ":=", "...",
	"<-", "->", "*", "&", "%", "+=", "-=", "*=", "/=",
}

var cppSymbols = []string{
	"{}", "[]", "()", "::", "->", "<<", ">>", "&&", "||", "!=", "==",
	"++", "--", "+=", "-=", "*=", "/=", "&", "*", "//", "/*", "*/", "#",
}

var jsSymbols = []string{
	"{}", "[]", "()", "=>", "===", "!==", "&&", "||", "??", "?.",
	"++", "--", "+=", "-=", "${", "`", "...", "//", "/*", "*/",
}

var rustSymbols = []string{
	"{}", "[]", "()", "::", "->", "=>", "..", "..=", "&&", "||",
	"==", "!=", "+=", "-=", "&", "*", "!", "//", "/*", "*/", "'a",
}

var pythonSymbols = []string{
	"[]", "()", "{}", "!=", "==", "<=", ">=", "and", "or", "not",
	"->", "**", "//", "+=", "-=", "*=", "/=", "@",
}

func (g *Generator) GenerateLesson(lessonType LessonType, length int) string {
	switch lessonType {
	case TypeBigrams:
		return g.generateFromList(commonBigrams, length, " ")
	case TypeWords:
		switch g.Language {
		case LangSpanish:
			return g.generateFromList(spanishWords, length, " ")
		case LangEnglish:
			return g.generateFromList(englishWords, length, " ")
		case LangFrench:
			return g.generateFromList(frenchWords, length, " ")
		case LangGerman:
			return g.generateFromList(germanWords, length, " ")
		case LangPython:
			return g.generateFromList(pythonKeywords, length, " ")
		case LangCpp:
			return g.generateFromList(cppKeywords, length, " ")
		case LangJavascript:
			return g.generateFromList(jsKeywords, length, " ")
		case LangRust:
			return g.generateFromList(rustKeywords, length, " ")
		default: // Go
			return g.generateFromList(goKeywords, length, " ")
		}
	case TypeSymbols:
		switch g.Language {
		case LangPython:
			return g.generateFromList(pythonSymbols, length, " ")
		case LangCpp:
			return g.generateFromList(cppSymbols, length, " ")
		case LangJavascript:
			return g.generateFromList(jsSymbols, length, " ")
		case LangRust:
			return g.generateFromList(rustSymbols, length, " ")
		default:
			return g.generateFromList(goSymbols, length, " ")
		}
	case TypeCode:
		switch g.Language {
		case LangPython:
			return g.generateCode(pythonSnippets, length)
		case LangCpp:
			return g.generateCode(cppSnippets, length)
		case LangJavascript:
			return g.generateCode(jsSnippets, length)
		case LangRust:
			return g.generateCode(rustSnippets, length)
		case LangSpanish, LangEnglish, LangFrench, LangGerman:
			// For natural languages, return longer word sequences
			var pool []string
			if g.Language == LangSpanish {
				pool = spanishWords
			} else if g.Language == LangFrench {
				pool = frenchWords
			} else if g.Language == LangGerman {
				pool = germanWords
			} else {
				pool = englishWords
			}
			return g.generateFromList(pool, length*2, " ")
		default: // Go
			return g.generateCode(goSnippets, length)
		}
	default:
		return g.generateFromList(goKeywords, length, " ")
	}
}

func (g *Generator) GenerateFromFile(filepath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (g *Generator) generateFromList(list []string, count int, sep string) string {
	var result []string
	for i := 0; i < count; i++ {
		result = append(result, list[g.rand.Intn(len(list))])
	}
	return strings.Join(result, sep)
}

func (g *Generator) generateCode(snippets []string, count int) string {
	var result []string
	for i := 0; i < count; i++ {
		result = append(result, snippets[g.rand.Intn(len(snippets))])
	}
	return strings.Join(result, "\n\n")
}

// GenerateWeaknessLesson creates a lesson focused on weak keys
func (g *Generator) GenerateWeaknessLesson(weakKeys []WeakKey, length int) string {
	if len(weakKeys) == 0 {
		return g.GenerateLesson(TypeWords, length)
	}

	// Determine word pool based on language
	var sourcePool []string
	switch g.Language {
	case LangSpanish:
		sourcePool = spanishWords
	case LangEnglish:
		sourcePool = englishWords
	case LangFrench:
		sourcePool = frenchWords
	case LangGerman:
		sourcePool = germanWords
	case LangPython:
		sourcePool = append(pythonKeywords, pythonSymbols...)
	case LangCpp:
		sourcePool = append(cppKeywords, cppSymbols...)
	case LangJavascript:
		sourcePool = append(jsKeywords, jsSymbols...)
	case LangRust:
		sourcePool = append(rustKeywords, rustSymbols...)
	default: // Go
		sourcePool = append(goKeywords, goSymbols...)
	}

	// Build a pool of words containing weak keys
	var wordPool []string
	seen := make(map[string]bool)

	// For each weak key, find words/bigrams that contain it
	for _, weak := range weakKeys {
		for _, word := range sourcePool {
			if strings.Contains(word, weak.Key) && !seen[word] {
				wordPool = append(wordPool, word)
				seen[word] = true
			}
		}
		
		// Always try bigrams too as fallback
		for _, bigram := range commonBigrams {
			if strings.Contains(bigram, weak.Key) && !seen[bigram] {
				wordPool = append(wordPool, bigram)
				seen[bigram] = true
			}
		}
	}

	// If pool is still empty or too small, add some random words
	if len(wordPool) < 5 {
		wordPool = append(wordPool, sourcePool...)
	}

	// Generate lesson
	var result []string
	for i := 0; i < length; i++ {
		result = append(result, wordPool[g.rand.Intn(len(wordPool))])
	}

	return strings.Join(result, " ")
}
