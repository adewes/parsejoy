package set

type SetError struct {
    msg string
}

func (e *SetError) Error() string {return e.msg}

type Set interface {
    Add(value interface{}) error
    Subtract(w Set) (Set, error)
    Union(w Set) (Set, error)
    Intersect(w Set) (Set, error)
    N() uint
    Remove(item interface{}) bool
    Contains(item interface{}) bool
}

type BitSet struct {
    v []uint64 //maximum of 640 distinct tokens    
    n uint
    Grammar *BitGrammar
}

type BitGrammar struct {
    n uint
    Mapping map[interface{}]uint
}

func (s *BitGrammar) Initialize() {
    s.n = 0
    s.Mapping = make(map[interface{}]uint)
}

func (s *BitGrammar) GetOrAdd(value interface{}) uint {
    existingId, ok := s.Mapping[value]
    if ok {
        return existingId
    }
    s.n+=1
    s.Mapping[value] = s.n
    return s.n
}

func (s *BitGrammar) ValueForId(id uint) interface{} {
    for key, value := range s.Mapping {
        if value == id {
            return key
        }
    }
    return nil
}

func (s *BitGrammar) Get(value interface{}) (uint, bool) {
    id, ok := s.Mapping[value]
    return id, ok
}

func (s *BitGrammar) AddAs(item interface{},id uint) bool {
    existingId, ok := s.Mapping[item]
    if ok {
        if existingId != id {
            return false
        }
        return true
    }
    for key, value := range s.Mapping {
        if key != item && value == id {
            //id is already assigned
            return false
        }
    }
    if id > s.n {
        s.n = id
    }
    s.Mapping[item] = id
    return true
}

func (s *BitSet) Reset() {
    for i := uint(0) ; i < s.n ; i++ {
        s.v[i] = 0
    }
}

func (s *BitSet) Initialize(bg *BitGrammar) {
    s.Grammar = bg
    if bg.n % 64 == 0 {
        s.n = uint(bg.n/64)
    } else {
        s.n = uint(bg.n/64)+1
    }
    s.v = make([]uint64,s.n)
}

func (s *BitSet) Add(item interface{}) error {
    id := s.Grammar.GetOrAdd(item)  
    s.add(id)
    return nil
}

func (s *BitSet) AddById(id uint) error {
    return s.add(id)
}

func (s *BitSet) add(id uint) error {
    var pos, offset uint
    pos = id / 64  
    offset = id % 64
    if pos >= s.n {
        oldV := s.v
        s.n = pos+1
        s.v = make([]uint64,s.n)
        copy(s.v,oldV)
    }
    s.v[pos]|= 1 << offset
    return nil
}

func (s *BitSet) Remove(item interface{}) bool {
    id, ok := s.Grammar.Get(item)
    if !ok {
        return false
    }
    pos := id / 64
    offset := id % 64
    s.v[pos]^= 1 << offset    
    return true
}

func (s *BitSet) Intersect (w Set) (Set, error) {
    wb, ok := w.(*BitSet)
    if ! ok {
        return s, &SetError{"Can only intersect BitSet with BitSet"}
    }
    h := BitSet{}
    h.Initialize(s.Grammar)
    if s.Grammar != wb.Grammar {
        return &h, &SetError{"grammars do not match"}
    }
    var i uint
    k := s.n
    if wb.n < k {
        k = wb.n
    }
    for i=0;i<k;i+=1 {
        h.v[i] = s.v[i] & wb.v[i]
    }
    return &h, nil
}

func (s *BitSet) Intersects (w Set) bool {
    wb, ok := w.(*BitSet)
    if ! ok {
        return false
    }
    if s.Grammar != wb.Grammar {
        return false
    }
    var i uint
    k := uint(s.n)
    if wb.n < k {
        k = wb.n
    }
    for i=0;i<k;i+=1 {
        if s.v[i] & wb.v[i] != 0 {
            return true
        }
    }
    return false
}

func (s *BitSet) Union (w Set) (Set, error) {

    wb, ok := w.(*BitSet)
    if ! ok {
        return s, &SetError{"Can only intersect BitSet with BitSet"}
    }

    h := BitSet{}
    h.Initialize(s.Grammar)
    if s.Grammar != wb.Grammar {
        return &h, &SetError{"grammars do not match"}
    }
    var i uint
    for i=0;i<h.n;i+=1 {
        if i < s.n && i < wb.n {
            h.v[i] = s.v[i] | wb.v[i]
        } else if i < s.n {
            h.v[i] = s.v[i]
        } else if i < wb.n {
            h.v[i] = wb.v[i]
        }
    }
    return &h, nil
}

func (s *BitSet) Subtract (w Set) (Set, error) {
    wb, ok := w.(*BitSet)
    if !ok {
        return s, &SetError{"Expected a bit set!"}
    }
    h := BitSet{}
    h.Initialize(s.Grammar)
    if s.Grammar != wb.Grammar {
        return &h, &SetError{"grammars do not match"}
    }
    var i uint
    for i=0;i<h.n;i+=1 {
        if i >= wb.n {
            h.v[i] = s.v[i]
        } else if i < s.n {
            h.v[i] = s.v[i] ^ (wb.v[i] & s.v[i])
        }
    }
    return &h, nil
}

func (s *BitSet) Contains (item interface{}) bool {
    id, ok := s.Grammar.Mapping[item]
    if !ok {
        return false
    }
    var pos, offset uint
    pos = id / 64
    offset = id % 64
    if s.v[pos] & (1 << offset) != 0 {
        return true
    }
    return false    
}

func (s *BitSet) N () uint {
    var i, j, cnt uint
    for i = 0; i < s.n;i++ {
        for j =0 ; j < 64 ; j++{
            if s.v[i] & (1 << j) != 0 {
                cnt += 1
            }
        }    
    }
    return cnt
}

func (s *BitSet) ContainsId (id uint) bool {
    var pos, offset uint
    pos = id / 64
    offset = id % 64
    return s.v[pos] & (1 << offset) != 0
}

func (s *BitSet) AsList() []interface{} {
    var pos, offset uint
    output := make([]interface{},0)
    for key, value := range s.Grammar.Mapping {
        pos = value / 64
        offset = value % 64
        if pos >= s.n {
            continue
        }
        if s.v[pos] & (1 << offset) != 0 {
            output = append(output, key)
        }
    }
    return output
}

type HashSet struct {
    Items map[interface{}]bool
}

func (s *HashSet) Initialize() {
    s.Items = make(map[interface{}]bool)
}


func (s *HashSet) Add (item interface{}) error {
    s.Items[item] = true
    return nil
}

func (s *HashSet) AsList() []interface{} {
    keys := make([]interface{},0,len(s.Items))
    for key := range s.Items {
        keys = append(keys,key)
    }
    return keys
}

func (s *HashSet) Intersect(w Set) (Set, error) {
    wh, ok := w.(*HashSet)
    if ! ok {
        return w,&SetError{"Expected hash set!"}
    }
    newSet := HashSet{make(map[interface{}]bool)}
    for key,_ := range s.Items {
        _, ok := wh.Items[key]
        if ok {
            newSet.Items[key] = true
        }
    }
    return &newSet, nil
}

func (s *HashSet) Union(w Set) (Set, error) {
    wh, ok := w.(*HashSet)
    if ! ok {
        return w,&SetError{"Expected hash set!"}
    }
    newSet := HashSet{make(map[interface{}]bool)}
    for key, _ := range s.Items {
        newSet.Items[key] = true
    }
    for key, _ := range wh.Items {
        newSet.Items[key] = true
    }
    return &newSet, nil
}

func (s *HashSet) Remove(item interface{}) bool {
    _ , ok := s.Items[item]
    if ok {
        delete(s.Items,item)
    } 
    return true
}

func (s *HashSet) Subtract(w Set) (Set, error) {
    wh, ok := w.(*HashSet)
    if ! ok {
        return w,&SetError{"Expected hash set!"}
    }
    newSet := HashSet{make(map[interface{}]bool)}
    for key, _ := range s.Items {
        _, ok := wh.Items[key]
        if !ok {
            newSet.Items[key] = true
        }
    }
    return &newSet, nil
}

func (s *HashSet) Contains(item interface{}) bool {
    _, ok := s.Items[item]
    return ok
}

func (s *HashSet) N() uint {
    return uint(len(s.Items))
}
