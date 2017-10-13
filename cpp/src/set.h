#pragma once

#include <unordered_set>
#include <unordered_map>
#include <memory>
#include <vector>
#include <string>
#include <boost/format.hpp>
#include <iostream>

using namespace std;

namespace sscientists{
namespace parsejoy{

template<typename T> class Set{
public:
    virtual bool Add(const T& element) = 0;
    virtual bool Remove(const T& element) = 0;
    virtual bool Contains(const T& element) = 0;
    virtual bool Intersects(const Set<T>& set) = 0;
    virtual unsigned int N() = 0;
};

template<typename T> class HashSet : public Set<T>{
public:
    ~HashSet();
    HashSet();
    bool Add(const T& element) override;
    bool Remove(const T& element) override;
    bool Contains(const T& element) override;
    HashSet<T> Subtract(const HashSet<T>& set);
    HashSet<T> Union(const HashSet<T>& set);
    HashSet<T> Intersect(const HashSet<T>& set);
    bool Intersects(const Set<T>& set) override;
    unsigned int N();
private:
    unordered_set<T> entries_;
};

template<typename T> class BitGrammar {
public:
    BitGrammar();
    ~BitGrammar();
    unsigned int GetOrAdd(const T& value);
    const T ValueForId(unsigned int id);
    unsigned int Get(const T& value);
    void AddAs(const T& value, unsigned int id);
    unsigned int N();
private:
    //we add hash<T> as a hashing function...
    unordered_map<const T,unsigned int, hash<T>> entries_;
    unsigned int n_;
};

typedef unsigned long long bitset_long;

template<typename T> class BitSet : public Set<T>{
public:
    BitSet(const shared_ptr<BitGrammar<T>> grammar);
    ~BitSet();
    void Reset();
    bool ContainsId(const unsigned int id);
    vector<T> AsVector();
    bool Add(const T& element) override;
    bool Remove(const T& element) override;
    bool Contains(const T& element) override;
    BitSet<T> Subtract(const BitSet<T>& set);
    BitSet<T> Union(const BitSet<T>& set);
    BitSet<T> Intersect(const BitSet<T>& set);
    bool Intersects(const Set<T>& set) override;
    unsigned int N();
protected:
    bool add(unsigned int id);
    void resize(unsigned int n);
    shared_ptr<BitGrammar<T>> grammar_;
    vector<bitset_long> entries_;
    unsigned int n_;
    static constexpr unsigned int width_ = sizeof(bitset_long)*8;
};


template<typename T>
HashSet<T>::HashSet(){}

template<typename T>
HashSet<T>::~HashSet(){}

template<typename T>
bool HashSet<T>::Add(const T& element){
    entries_.insert(element);
    return true;
}

template<typename T>
bool HashSet<T>::Remove(const T& element){
    entries_.erase(element);
    return true;
}

template<typename T>
bool HashSet<T>::Contains(const T& element){
    return entries_.find(element) != entries_.end();
}

template<typename T>
HashSet<T> HashSet<T>::Subtract(const HashSet<T>& set){
    auto hashSet = *this;
    for(auto it=set.entries_.begin();it!=set.entries_.end();++it)
    {
        auto existingIt = hashSet.entries_.find(*it);
        if (existingIt != hashSet.entries_.end())
            hashSet.entries_.erase(existingIt);
    }
    return hashSet;
}

template<typename T>
HashSet<T> HashSet<T>::Union(const HashSet<T>& set){
    auto hashSet = *this;
    for(auto it=set.entries_.begin();it!=set.entries_.end();++it)
        hashSet.entries_.insert(*it);
    return hashSet;
}

template<typename T>
HashSet<T> HashSet<T>::Intersect(const HashSet<T>& set){
    auto hashSet = *this;
    for(auto it=hashSet.entries_.begin();it!=hashSet.entries_.end();++it){
        auto existingIt = set.entries_.find(*it);
        if (existingIt == set.entries_.end())
            hashSet.entries_.erase(existingIt);
    }
    return hashSet;
}

template<typename T>
bool HashSet<T>::Intersects(const Set<T>& set){
    const HashSet& hashSet = static_cast<const HashSet<T>&>(set);
    for(auto it=entries_.begin();it!=entries_.end();++it){
        auto existingIt = hashSet.entries_.find(*it);
        if (existingIt != hashSet.entries_.end())
            return true;
    }
    return false;
}

template<typename T>
unsigned int HashSet<T>::N(){
    return entries_.size();
}

/*BitSet*/

template<typename T>
BitSet<T>::BitSet(const shared_ptr<BitGrammar<T>> grammar) : grammar_(grammar) {
    n_ = 0;
    unsigned int n = (unsigned int)(grammar_->N() / width_);
    if (grammar_->N()%width_ != 0)
        n++;
    resize(n);
}

template<typename T>
BitSet<T>::~BitSet(){
    cout << "Destroying bit set\n";
}

template<typename T>
void BitSet<T>::Reset(){
    for(int i=0;i < n_;i++)
        entries_[i] = 0;
}

template<typename T>
bool BitSet<T>::Add(const T& element){
    auto id = grammar_->GetOrAdd(element) - 1;
    return add(id);
}

template<typename T>
void BitSet<T>::resize(unsigned int n){
    auto oldN = n_;
    n_ = n;
    entries_.resize(n_, 0);
}

template<typename T>
bool BitSet<T>::add(unsigned int id){
    unsigned int pos, offset;
    pos = id / width_;
    offset = id % width_;
    if (pos >= n_){
        resize(pos + 1);
    }
    if (entries_[pos] & bitset_long(1) << offset)
        return false;
    entries_[pos] |= bitset_long(1) << offset;
    return true;
}

template<typename T>
bool BitSet<T>::Remove(const T& element){
    auto id = grammar_->Get(element) - 1;
    if (id == -1)
        return false;
    auto pos = id / width_;
    auto offset = id % width_;
    entries_[pos] ^= bitset_long(1) << offset;
    return true;
}

template<typename T>
bool BitSet<T>::Contains(const T& element){
    auto id = grammar_->Get(element);
    if (id == 0)
        return false;
    return ContainsId(id);
}

template<typename T>
bool BitSet<T>::ContainsId(unsigned int id){
    auto pos = (id-1) / width_;
    auto offset = (id-1) % width_;
    if (pos >= n_)
        return false;
    return (entries_[pos] & (bitset_long(1)<<offset)) != 0;
}

template<typename T>
BitSet<T> BitSet<T>::Subtract(const BitSet<T>& set){
    if (set.grammar_ != grammar_)
        throw runtime_error("Grammars do not match!");
    auto newBitSet = *this;
    for(unsigned int i=0; i<newBitSet.n_; i++){
        if (i < set.n_)
            newBitSet.entries_[i] ^= (newBitSet.entries_[i] & set.entries_[i]);
    }
    return newBitSet;
}

template<typename T>
BitSet<T> BitSet<T>::Union(const BitSet<T>& set){
    if (set.grammar_ != grammar_)
        throw runtime_error("Grammars do not match!");
    auto newBitSet = *this;
    if (newBitSet.n_ < set.n_)
        newBitSet.resize(set.n_);
    for(unsigned int i=0; i<newBitSet.n_; i++){
        if (i < set.n_)
            newBitSet.entries_[i] |= set.entries_[i];
    }
    return newBitSet;
}

template<typename T>
BitSet<T> BitSet<T>::Intersect(const BitSet<T>& set){
    if (set.grammar_ != grammar_)
        throw runtime_error("Grammars do not match!");
    auto newBitSet = *this;
    if (newBitSet.n_ > set.n_)
        newBitSet.resize(set.n_);
    for(unsigned int i=0; i<newBitSet.n_; i++){
        if (i < set.n_)
            newBitSet.entries_[i] &= set.entries_[i];
    }
    return newBitSet;
}

template<typename T>
bool BitSet<T>::Intersects(const Set<T>& set){
    const BitSet& bitSet = static_cast<const BitSet<T>&>(set);
    if (bitSet.grammar_ != grammar_)
        throw runtime_error("Grammars do not match!");
    for(unsigned int i=0; i<min(n_,bitSet.n_); i++){
        if (entries_[i] & bitSet.entries_[i])
            return true;
    }
    return false;
}

template<typename T>
unsigned int BitSet<T>::N(){
    unsigned int cnt = 0;
    for(unsigned int i=0; i<n_; i++)
        for(unsigned int j=0; j<width_; j++)
            if (((entries_[i] >> j) & bitset_long(1)) == 1)
                cnt++;
    return cnt;
}

template<typename T>
vector<T> BitSet<T>::AsVector(){
    //warning, this function is quite inefficient and intended only for debugging purposes...
    vector<T> values;
    for(unsigned int i=0; i<n_*width_; i++){
        unsigned int pos = i / width_;
        unsigned int offset = i % width_;
        if (entries_[pos] >> offset & bitset_long(1) == 1)
            values.push_back(grammar_->ValueForId(i+1));
    }
    return values;
}

/*BitGrammar*/

template<typename T>
BitGrammar<T>::BitGrammar(){
    n_ = 0;
}

template<typename T>
BitGrammar<T>::~BitGrammar(){
    cout << "Destroying bit grammar...\n";
}

template<typename T>
unsigned int BitGrammar<T>::N(){
    return n_;
}

template<typename T>
unsigned int BitGrammar<T>::GetOrAdd(const T& value){
    auto existingId = entries_.find(value);
    if (existingId == entries_.end()){
        n_ += 1;
        entries_.insert(make_pair(value,n_));
        return n_;
    } else
        return existingId->second;
}

template<typename T>
const T BitGrammar<T>::ValueForId(unsigned int id){
    for(auto it=entries_.begin();it != entries_.end();++it){
        if (it->second == id)
            return it->first;
    }
    throw runtime_error(boost::str(boost::format("No value for ID %1%") % id ));
}

template<typename T>
unsigned int BitGrammar<T>::Get(const T& value){
    auto it = entries_.find(value);
    if (it != entries_.end())
        return it->second;
    return 0;
}

template<typename T>
void BitGrammar<T>::AddAs(const T& value, unsigned int id){
    auto existingId = Get(value);
    for(auto it=entries_.begin();it != entries_.end();++it){
        if (it->second == id){
            if (it->first == value)
                return;//the ID already exists, but it's the same element so it's okay.
            throw runtime_error(boost::str(boost::format("ID %1% already exists!") % id ));
        }
    }
    entries_.insert(make_pair(value, id));
}


}
}
