namespace std {

  %feature("novaluewrapper") unique_ptr;
  template <typename Type>
  struct unique_ptr {
     typedef Type* pointer;

     explicit unique_ptr( pointer Ptr );
     unique_ptr (unique_ptr&& Right);
     template<class Type2, Class Del2> unique_ptr( unique_ptr<Type2, Del2>&& Right );
     unique_ptr( const unique_ptr& Right) = delete;


     pointer operator-> () const;
     pointer get () const;

     ~unique_ptr();
  };
}

%define wrap_unique_ptr(Name, Type)
  %template(Name) std::unique_ptr<Type>;
  %newobject std::unique_ptr<Type>::release;

  %typemap(out) std::unique_ptr<Type> %{
    $result = SWIG_NewPointerObj(L, new $1_ltype(std::move($1)), $&1_descriptor, SWIG_POINTER_OWN);
    SWIG_arg++;
  %}

%enddef 

namespace std {

template<class T>
class shared_ptr
{
    public:
        T *     operator-> () const;
        void    reset();
        T *     get() const;
};

%define wrap_shared_ptr(Name, Type)
  %template(Name) std::shared_ptr<Type>;
  //%newobject std::shared_ptr<Type>::release;

  %typemap(out) std::shared_ptr<Type> %{
    SWIG_NewPointerObj(L, new $1_ltype($1), $&1_descriptor, SWIG_POINTER_OWN);
    SWIG_arg++;
  %}

%enddef

}
