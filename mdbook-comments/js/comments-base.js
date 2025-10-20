"use strict";(()=>{var V,h,ge,Oe,M,pe,ve,be,Ce,ee,Y,Z,We,R={},xe=[],je=/acit|ex(?:s|g|n|p|$)|rph|grid|ows|mnc|ntw|ine[ch]|zoo|^ord|itera/i,J=Array.isArray;function S(e,t){for(var n in t)e[n]=t[n];return e}function te(e){e&&e.parentNode&&e.parentNode.removeChild(e)}function Ve(e,t,n){var o,a,r,i={};for(r in t)r=="key"?o=t[r]:r=="ref"?a=t[r]:i[r]=t[r];if(arguments.length>2&&(i.children=arguments.length>3?V.call(arguments,2):n),typeof e=="function"&&e.defaultProps!=null)for(r in e.defaultProps)i[r]===void 0&&(i[r]=e.defaultProps[r]);return O(e,i,o,a,null)}function O(e,t,n,o,a){var r={type:e,props:t,key:n,ref:o,__k:null,__:null,__b:0,__e:null,__c:null,constructor:void 0,__v:a??++ge,__i:-1,__u:0};return a==null&&h.vnode!=null&&h.vnode(r),r}function F(e){return e.children}function W(e,t){this.props=e,this.context=t}function H(e,t){if(t==null)return e.__?H(e.__,e.__i+1):null;for(var n;t<e.__k.length;t++)if((n=e.__k[t])!=null&&n.__e!=null)return n.__e;return typeof e.type=="function"?H(e):null}function ke(e){var t,n;if((e=e.__)!=null&&e.__c!=null){for(e.__e=e.__c.base=null,t=0;t<e.__k.length;t++)if((n=e.__k[t])!=null&&n.__e!=null){e.__e=e.__c.base=n.__e;break}return ke(e)}}function fe(e){(!e.__d&&(e.__d=!0)&&M.push(e)&&!j.__r++||pe!=h.debounceRendering)&&((pe=h.debounceRendering)||ve)(j)}function j(){for(var e,t,n,o,a,r,i,l=1;M.length;)M.length>l&&M.sort(be),e=M.shift(),l=M.length,e.__d&&(n=void 0,o=void 0,a=(o=(t=e).__v).__e,r=[],i=[],t.__P&&((n=S({},o)).__v=o.__v+1,h.vnode&&h.vnode(n),ne(t.__P,n,o,t.__n,t.__P.namespaceURI,32&o.__u?[a]:null,r,a??H(o),!!(32&o.__u),i),n.__v=o.__v,n.__.__k[n.__i]=n,Ae(r,n,i),o.__e=o.__=null,n.__e!=a&&ke(n)));j.__r=0}function we(e,t,n,o,a,r,i,l,_,c,p){var s,f,u,C,k,x,y,g=o&&o.__k||xe,P=t.length;for(_=Je(n,t,g,_,P),s=0;s<P;s++)(u=n.__k[s])!=null&&(f=u.__i==-1?R:g[u.__i]||R,u.__i=s,x=ne(e,u,f,a,r,i,l,_,c,p),C=u.__e,u.ref&&f.ref!=u.ref&&(f.ref&&oe(f.ref,null,u),p.push(u.ref,u.__c||C,u)),k==null&&C!=null&&(k=C),(y=!!(4&u.__u))||f.__k===u.__k?_=Se(u,_,e,y):typeof u.type=="function"&&x!==void 0?_=x:C&&(_=C.nextSibling),u.__u&=-7);return n.__e=k,_}function Je(e,t,n,o,a){var r,i,l,_,c,p=n.length,s=p,f=0;for(e.__k=new Array(a),r=0;r<a;r++)(i=t[r])!=null&&typeof i!="boolean"&&typeof i!="function"?(_=r+f,(i=e.__k[r]=typeof i=="string"||typeof i=="number"||typeof i=="bigint"||i.constructor==String?O(null,i,null,null,null):J(i)?O(F,{children:i},null,null,null):i.constructor==null&&i.__b>0?O(i.type,i.props,i.key,i.ref?i.ref:null,i.__v):i).__=e,i.__b=e.__b+1,l=null,(c=i.__i=Ge(i,n,_,s))!=-1&&(s--,(l=n[c])&&(l.__u|=2)),l==null||l.__v==null?(c==-1&&(a>p?f--:a<p&&f++),typeof i.type!="function"&&(i.__u|=4)):c!=_&&(c==_-1?f--:c==_+1?f++:(c>_?f--:f++,i.__u|=4))):e.__k[r]=null;if(s)for(r=0;r<p;r++)(l=n[r])!=null&&(2&l.__u)==0&&(l.__e==o&&(o=H(l)),Ee(l,l));return o}function Se(e,t,n,o){var a,r;if(typeof e.type=="function"){for(a=e.__k,r=0;a&&r<a.length;r++)a[r]&&(a[r].__=e,t=Se(a[r],t,n,o));return t}e.__e!=t&&(o&&(t&&e.type&&!t.parentNode&&(t=H(e)),n.insertBefore(e.__e,t||null)),t=e.__e);do t=t&&t.nextSibling;while(t!=null&&t.nodeType==8);return t}function Ge(e,t,n,o){var a,r,i,l=e.key,_=e.type,c=t[n],p=c!=null&&(2&c.__u)==0;if(c===null&&e.key==null||p&&l==c.key&&_==c.type)return n;if(o>(p?1:0)){for(a=n-1,r=n+1;a>=0||r<t.length;)if((c=t[i=a>=0?a--:r++])!=null&&(2&c.__u)==0&&l==c.key&&_==c.type)return i}return-1}function he(e,t,n){t[0]=="-"?e.setProperty(t,n??""):e[t]=n==null?"":typeof n!="number"||je.test(t)?n:n+"px"}function q(e,t,n,o,a){var r,i;e:if(t=="style")if(typeof n=="string")e.style.cssText=n;else{if(typeof o=="string"&&(e.style.cssText=o=""),o)for(t in o)n&&t in n||he(e.style,t,"");if(n)for(t in n)o&&n[t]==o[t]||he(e.style,t,n[t])}else if(t[0]=="o"&&t[1]=="n")r=t!=(t=t.replace(Ce,"$1")),i=t.toLowerCase(),t=i in e||t=="onFocusOut"||t=="onFocusIn"?i.slice(2):t.slice(2),e.l||(e.l={}),e.l[t+r]=n,n?o?n.u=o.u:(n.u=ee,e.addEventListener(t,r?Z:Y,r)):e.removeEventListener(t,r?Z:Y,r);else{if(a=="http://www.w3.org/2000/svg")t=t.replace(/xlink(H|:h)/,"h").replace(/sName$/,"s");else if(t!="width"&&t!="height"&&t!="href"&&t!="list"&&t!="form"&&t!="tabIndex"&&t!="download"&&t!="rowSpan"&&t!="colSpan"&&t!="role"&&t!="popover"&&t in e)try{e[t]=n??"";break e}catch{}typeof n=="function"||(n==null||n===!1&&t[4]!="-"?e.removeAttribute(t):e.setAttribute(t,t=="popover"&&n==1?"":n))}}function ye(e){return function(t){if(this.l){var n=this.l[t.type+e];if(t.t==null)t.t=ee++;else if(t.t<n.u)return;return n(h.event?h.event(t):t)}}}function ne(e,t,n,o,a,r,i,l,_,c){var p,s,f,u,C,k,x,y,g,P,E,U,L,de,z,D,Q,w=t.type;if(t.constructor!=null)return null;128&n.__u&&(_=!!(32&n.__u),r=[l=t.__e=n.__e]),(p=h.__b)&&p(t);e:if(typeof w=="function")try{if(y=t.props,g="prototype"in w&&w.prototype.render,P=(p=w.contextType)&&o[p.__c],E=p?P?P.props.value:p.__:o,n.__c?x=(s=t.__c=n.__c).__=s.__E:(g?t.__c=s=new w(y,E):(t.__c=s=new W(y,E),s.constructor=w,s.render=Qe),P&&P.sub(s),s.props=y,s.state||(s.state={}),s.context=E,s.__n=o,f=s.__d=!0,s.__h=[],s._sb=[]),g&&s.__s==null&&(s.__s=s.state),g&&w.getDerivedStateFromProps!=null&&(s.__s==s.state&&(s.__s=S({},s.__s)),S(s.__s,w.getDerivedStateFromProps(y,s.__s))),u=s.props,C=s.state,s.__v=t,f)g&&w.getDerivedStateFromProps==null&&s.componentWillMount!=null&&s.componentWillMount(),g&&s.componentDidMount!=null&&s.__h.push(s.componentDidMount);else{if(g&&w.getDerivedStateFromProps==null&&y!==u&&s.componentWillReceiveProps!=null&&s.componentWillReceiveProps(y,E),!s.__e&&s.shouldComponentUpdate!=null&&s.shouldComponentUpdate(y,s.__s,E)===!1||t.__v==n.__v){for(t.__v!=n.__v&&(s.props=y,s.state=s.__s,s.__d=!1),t.__e=n.__e,t.__k=n.__k,t.__k.some(function(T){T&&(T.__=t)}),U=0;U<s._sb.length;U++)s.__h.push(s._sb[U]);s._sb=[],s.__h.length&&i.push(s);break e}s.componentWillUpdate!=null&&s.componentWillUpdate(y,s.__s,E),g&&s.componentDidUpdate!=null&&s.__h.push(function(){s.componentDidUpdate(u,C,k)})}if(s.context=E,s.props=y,s.__P=e,s.__e=!1,L=h.__r,de=0,g){for(s.state=s.__s,s.__d=!1,L&&L(t),p=s.render(s.props,s.state,s.context),z=0;z<s._sb.length;z++)s.__h.push(s._sb[z]);s._sb=[]}else do s.__d=!1,L&&L(t),p=s.render(s.props,s.state,s.context),s.state=s.__s;while(s.__d&&++de<25);s.state=s.__s,s.getChildContext!=null&&(o=S(S({},o),s.getChildContext())),g&&!f&&s.getSnapshotBeforeUpdate!=null&&(k=s.getSnapshotBeforeUpdate(u,C)),D=p,p!=null&&p.type===F&&p.key==null&&(D=Pe(p.props.children)),l=we(e,J(D)?D:[D],t,n,o,a,r,i,l,_,c),s.base=t.__e,t.__u&=-161,s.__h.length&&i.push(s),x&&(s.__E=s.__=null)}catch(T){if(t.__v=null,_||r!=null)if(T.then){for(t.__u|=_?160:128;l&&l.nodeType==8&&l.nextSibling;)l=l.nextSibling;r[r.indexOf(l)]=null,t.__e=l}else{for(Q=r.length;Q--;)te(r[Q]);X(t)}else t.__e=n.__e,t.__k=n.__k,T.then||X(t);h.__e(T,t,n)}else r==null&&t.__v==n.__v?(t.__k=n.__k,t.__e=n.__e):l=t.__e=Ke(n.__e,t,n,o,a,r,i,_,c);return(p=h.diffed)&&p(t),128&t.__u?void 0:l}function X(e){e&&e.__c&&(e.__c.__e=!0),e&&e.__k&&e.__k.forEach(X)}function Ae(e,t,n){for(var o=0;o<n.length;o++)oe(n[o],n[++o],n[++o]);h.__c&&h.__c(t,e),e.some(function(a){try{e=a.__h,a.__h=[],e.some(function(r){r.call(a)})}catch(r){h.__e(r,a.__v)}})}function Pe(e){return typeof e!="object"||e==null||e.__b&&e.__b>0?e:J(e)?e.map(Pe):S({},e)}function Ke(e,t,n,o,a,r,i,l,_){var c,p,s,f,u,C,k,x=n.props,y=t.props,g=t.type;if(g=="svg"?a="http://www.w3.org/2000/svg":g=="math"?a="http://www.w3.org/1998/Math/MathML":a||(a="http://www.w3.org/1999/xhtml"),r!=null){for(c=0;c<r.length;c++)if((u=r[c])&&"setAttribute"in u==!!g&&(g?u.localName==g:u.nodeType==3)){e=u,r[c]=null;break}}if(e==null){if(g==null)return document.createTextNode(y);e=document.createElementNS(a,g,y.is&&y),l&&(h.__m&&h.__m(t,r),l=!1),r=null}if(g==null)x===y||l&&e.data==y||(e.data=y);else{if(r=r&&V.call(e.childNodes),x=n.props||R,!l&&r!=null)for(x={},c=0;c<e.attributes.length;c++)x[(u=e.attributes[c]).name]=u.value;for(c in x)if(u=x[c],c!="children"){if(c=="dangerouslySetInnerHTML")s=u;else if(!(c in y)){if(c=="value"&&"defaultValue"in y||c=="checked"&&"defaultChecked"in y)continue;q(e,c,null,u,a)}}for(c in y)u=y[c],c=="children"?f=u:c=="dangerouslySetInnerHTML"?p=u:c=="value"?C=u:c=="checked"?k=u:l&&typeof u!="function"||x[c]===u||q(e,c,u,x[c],a);if(p)l||s&&(p.__html==s.__html||p.__html==e.innerHTML)||(e.innerHTML=p.__html),t.__k=[];else if(s&&(e.innerHTML=""),we(t.type=="template"?e.content:e,J(f)?f:[f],t,n,o,g=="foreignObject"?"http://www.w3.org/1999/xhtml":a,r,i,r?r[0]:n.__k&&H(n,0),l,_),r!=null)for(c=r.length;c--;)te(r[c]);l||(c="value",g=="progress"&&C==null?e.removeAttribute("value"):C!=null&&(C!==e[c]||g=="progress"&&!C||g=="option"&&C!=x[c])&&q(e,c,C,x[c],a),c="checked",k!=null&&k!=e[c]&&q(e,c,k,x[c],a))}return e}function oe(e,t,n){try{if(typeof e=="function"){var o=typeof e.__u=="function";o&&e.__u(),o&&t==null||(e.__u=e(t))}else e.current=t}catch(a){h.__e(a,n)}}function Ee(e,t,n){var o,a;if(h.unmount&&h.unmount(e),(o=e.ref)&&(o.current&&o.current!=e.__e||oe(o,null,t)),(o=e.__c)!=null){if(o.componentWillUnmount)try{o.componentWillUnmount()}catch(r){h.__e(r,t)}o.base=o.__P=null}if(o=e.__k)for(a=0;a<o.length;a++)o[a]&&Ee(o[a],t,n||typeof e.type!="function");n||te(e.__e),e.__c=e.__=e.__e=void 0}function Qe(e,t,n){return this.constructor(e,n)}function $(e,t,n){var o,a,r,i;t==document&&(t=document.documentElement),h.__&&h.__(e,t),a=(o=typeof n=="function")?null:n&&n.__k||t.__k,r=[],i=[],ne(t,e=(!o&&n||t).__k=Ve(F,null,[e]),a||R,R,t.namespaceURI,!o&&n?[n]:a?null:t.firstChild?V.call(t.childNodes):null,r,!o&&n?n:a?a.__e:t.firstChild,o,i),Ae(r,e,i)}V=xe.slice,h={__e:function(e,t,n,o){for(var a,r,i;t=t.__;)if((a=t.__c)&&!a.__)try{if((r=a.constructor)&&r.getDerivedStateFromError!=null&&(a.setState(r.getDerivedStateFromError(e)),i=a.__d),a.componentDidCatch!=null&&(a.componentDidCatch(e,o||{}),i=a.__d),i)return a.__E=a}catch(l){e=l}throw e}},ge=0,Oe=function(e){return e!=null&&e.constructor==null},W.prototype.setState=function(e,t){var n;n=this.__s!=null&&this.__s!=this.state?this.__s:this.__s=S({},this.state),typeof e=="function"&&(e=e(S({},n),this.props)),e&&S(n,e),e!=null&&this.__v&&(t&&this._sb.push(t),fe(this))},W.prototype.forceUpdate=function(e){this.__v&&(this.__e=!0,e&&this.__h.push(e),fe(this))},W.prototype.render=F,M=[],ve=typeof Promise=="function"?Promise.prototype.then.bind(Promise.resolve()):setTimeout,be=function(e,t){return e.__v.__b-t.__v.__b},j.__r=0,Ce=/(PointerCapture)$|Capture$/i,ee=0,Y=ye(!1),Z=ye(!0),We=0;var ae,v,re,Me,ie=0,$e=[],b=h,Te=b.__b,He=b.__r,Fe=b.diffed,Ie=b.__c,Le=b.unmount,De=b.__;function Ye(e,t){b.__h&&b.__h(v,e,ie||t),ie=0;var n=v.__H||(v.__H={__:[],__h:[]});return e>=n.__.length&&n.__.push({}),n.__[e]}function A(e){return ie=1,Ze(Be,e)}function Ze(e,t,n){var o=Ye(ae++,2);if(o.t=e,!o.__c&&(o.__=[n?n(t):Be(void 0,t),function(l){var _=o.__N?o.__N[0]:o.__[0],c=o.t(_,l);_!==c&&(o.__N=[c,o.__[1]],o.__c.setState({}))}],o.__c=v,!v.__f)){var a=function(l,_,c){if(!o.__c.__H)return!0;var p=o.__c.__H.__.filter(function(f){return!!f.__c});if(p.every(function(f){return!f.__N}))return!r||r.call(this,l,_,c);var s=o.__c.props!==l;return p.forEach(function(f){if(f.__N){var u=f.__[0];f.__=f.__N,f.__N=void 0,u!==f.__[0]&&(s=!0)}}),r&&r.call(this,l,_,c)||s};v.__f=!0;var r=v.shouldComponentUpdate,i=v.componentWillUpdate;v.componentWillUpdate=function(l,_,c){if(this.__e){var p=r;r=void 0,a(l,_,c),r=p}i&&i.call(this,l,_,c)},v.shouldComponentUpdate=a}return o.__N||o.__}function Xe(){for(var e;e=$e.shift();)if(e.__P&&e.__H)try{e.__H.__h.forEach(G),e.__H.__h.forEach(se),e.__H.__h=[]}catch(t){e.__H.__h=[],b.__e(t,e.__v)}}b.__b=function(e){v=null,Te&&Te(e)},b.__=function(e,t){e&&t.__k&&t.__k.__m&&(e.__m=t.__k.__m),De&&De(e,t)},b.__r=function(e){He&&He(e),ae=0;var t=(v=e.__c).__H;t&&(re===v?(t.__h=[],v.__h=[],t.__.forEach(function(n){n.__N&&(n.__=n.__N),n.u=n.__N=void 0})):(t.__h.forEach(G),t.__h.forEach(se),t.__h=[],ae=0)),re=v},b.diffed=function(e){Fe&&Fe(e);var t=e.__c;t&&t.__H&&(t.__H.__h.length&&($e.push(t)!==1&&Me===b.requestAnimationFrame||((Me=b.requestAnimationFrame)||et)(Xe)),t.__H.__.forEach(function(n){n.u&&(n.__H=n.u),n.u=void 0})),re=v=null},b.__c=function(e,t){t.some(function(n){try{n.__h.forEach(G),n.__h=n.__h.filter(function(o){return!o.__||se(o)})}catch(o){t.some(function(a){a.__h&&(a.__h=[])}),t=[],b.__e(o,n.__v)}}),Ie&&Ie(e,t)},b.unmount=function(e){Le&&Le(e);var t,n=e.__c;n&&n.__H&&(n.__H.__.forEach(function(o){try{G(o)}catch(a){t=a}}),n.__H=void 0,t&&b.__e(t,n.__v))};var Re=typeof requestAnimationFrame=="function";function et(e){var t,n=function(){clearTimeout(o),Re&&cancelAnimationFrame(t),setTimeout(e)},o=setTimeout(n,35);Re&&(t=requestAnimationFrame(n))}function G(e){var t=v,n=e.__c;typeof n=="function"&&(e.__c=void 0,n()),v=t}function se(e){var t=v;e.__c=e.__(),v=t}function Be(e,t){return typeof t=="function"?t(e):t}var tt=0,Ct=Array.isArray;function m(e,t,n,o,a,r){t||(t={});var i,l,_=t;if("ref"in _)for(l in _={},t)l=="ref"?i=t[l]:_[l]=t[l];var c={type:e,props:_,key:n,ref:i,__k:null,__:null,__b:0,__e:null,__c:null,constructor:void 0,__v:--tt,__i:-1,__u:0,__source:a,__self:r};if(typeof e=="function"&&(i=e.defaultProps))for(l in i)_[l]===void 0&&(_[l]=i[l]);return h.vnode&&h.vnode(c),c}function ce({parentCommentId:e,backend:t,onReplySubmitted:n,onCancel:o}){let[a,r]=A(""),[i,l]=A(!1);return m("div",{class:"reply-form",children:m("form",{onSubmit:async c=>{if(c.preventDefault(),!a.trim()){alert("Please enter a reply");return}l(!0);try{let p=t.getCurrentAuthor?.()||"Anonymous";await t.saveReply(e,a.trim(),p),r(""),n()}catch(p){console.error("Error posting reply:",p),alert("Failed to post reply. Please try again.")}finally{l(!1)}},children:[m("textarea",{class:"reply-input",name:"reply-text",placeholder:"Write a reply...",rows:2,value:a,onInput:c=>r(c.target.value),disabled:i,autoFocus:!0}),m("div",{class:"reply-form-actions",children:[m("button",{type:"submit",class:"reply-submit",disabled:i,children:i?"Posting...":"Post Reply"}),m("button",{type:"button",class:"reply-cancel",onClick:o,disabled:i,children:"Cancel"})]})]})})}function K(e){let t=document.createElement("div");return t.textContent=e,t.innerHTML}function Ne(e){let t=new Date(e),o=new Date().getTime()-t.getTime(),a=Math.floor(o/6e4),r=Math.floor(o/36e5),i=Math.floor(o/864e5);return a<1?"just now":a<60?`${a} minute${a>1?"s":""} ago`:r<24?`${r} hour${r>1?"s":""} ago`:i<7?`${i} day${i>1?"s":""} ago`:t.toLocaleDateString()}function le({comment:e,backend:t,onUpdate:n}){let[o,a]=A(!1),r=()=>{a(!1),n()};return m("div",{class:"comment-item","data-comment-id":e.id,children:[m("div",{class:"comment-header",children:[m("span",{class:"comment-author",children:K(e.author||"Anonymous")}),m("span",{class:"comment-date",children:Ne(e.created)})]}),m("div",{class:"comment-text",dangerouslySetInnerHTML:{__html:K(e.text)}}),e.replies&&e.replies.length>0&&m("div",{class:"comment-replies",children:e.replies.map(i=>m("div",{class:"reply-item",children:[m("div",{class:"reply-header",children:[m("span",{class:"reply-author",children:K(i.author||"Anonymous")}),m("span",{class:"reply-date",children:Ne(i.created)})]}),m("div",{class:"reply-text",dangerouslySetInnerHTML:{__html:K(i.text)}})]},i.id))}),m("button",{class:"comment-reply-btn",onClick:()=>a(!o),children:o?"Cancel":"Reply"}),o&&m(ce,{parentCommentId:e.id,backend:t,onReplySubmitted:r,onCancel:()=>a(!1)})]})}function me({paragraphId:e,metadata:t,backend:n,onCommentSubmitted:o}){let[a,r]=A(""),[i,l]=A(n.getCurrentAuthor?.()||""),[_,c]=A(!1),p=n.showAuthorInput&&!n.getCurrentAuthor?.();return m("div",{class:"comment-form",children:m("form",{onSubmit:async f=>{if(f.preventDefault(),!a.trim()){alert("Please enter a comment");return}if(p&&!i.trim()){alert("Please enter your name");return}c(!0);try{i&&n.setCurrentAuthor&&n.setCurrentAuthor(i),await n.saveComment(e,t,a.trim(),i.trim()||"Anonymous"),r(""),o()}catch(u){console.error("Error posting comment:",u),alert("Failed to post comment. Please try again.")}finally{c(!1)}},children:[p&&m("input",{type:"text",class:"author-input",name:"author",placeholder:"Your name",value:i,onInput:f=>{let u=f.target.value;l(u),u&&n.setCurrentAuthor&&n.setCurrentAuthor(u)},disabled:_}),m("textarea",{class:"comment-input",name:"comment-text",placeholder:"Add a comment...",rows:3,value:a,onInput:f=>r(f.target.value),disabled:_}),m("button",{type:"submit",class:"comment-submit",disabled:_,children:_?"Submitting...":"Submit"})]})})}function B({paragraphId:e,metadata:t,comments:n,backend:o,onUpdate:a}){let r=n.filter(i=>i.paragraphId===e).map(i=>i.comment);return m("div",{id:`comments-${e}`,class:"comment-section","data-paragraph-id":e,children:[m("div",{class:"comment-list",children:r.length>0?r.map(i=>m(le,{comment:i,backend:o,onUpdate:a},i.id)):m("p",{class:"no-comments",children:"No comments yet. Be the first to comment!"})}),m(me,{paragraphId:e,metadata:t,backend:o,onCommentSubmitted:a})]})}function N(e){let t=document.createElement("div");return t.textContent=e,t.innerHTML}function Ue(e){let t=new Date(e),o=new Date().getTime()-t.getTime(),a=Math.floor(o/6e4),r=Math.floor(o/36e5),i=Math.floor(o/864e5);return a<1?"just now":a<60?`${a} minute${a>1?"s":""} ago`:r<24?`${r} hour${r>1?"s":""} ago`:i<7?`${i} day${i>1?"s":""} ago`:t.toLocaleDateString()}function nt({comment:e}){let t=e.metadata||{},n=t.context||{"heading-path":[]},o=t.content||"[Content not available]",a=n["heading-path"]||[];return m("div",{class:"orphaned-comment",children:[m("div",{class:"orphaned-comment-context",children:[m("strong",{children:"Original paragraph:"}),m("blockquote",{dangerouslySetInnerHTML:{__html:N(o)}}),a.length>0&&m("div",{class:"orphaned-comment-location",children:["Section: ",a.join(" > ")]})]}),m("div",{class:"comment-item",children:[m("div",{class:"comment-header",children:[m("span",{class:"comment-author",children:N(e.author||"Anonymous")}),m("span",{class:"comment-date",children:Ue(e.created)})]}),m("div",{class:"comment-text",dangerouslySetInnerHTML:{__html:N(e.text)}}),e.replies&&e.replies.length>0&&m("div",{class:"comment-replies",children:e.replies.map(r=>m("div",{class:"reply-item",children:[m("div",{class:"reply-header",children:[m("span",{class:"reply-author",children:N(r.author||"Anonymous")}),m("span",{class:"reply-date",children:Ue(r.created)})]}),m("div",{class:"reply-text",dangerouslySetInnerHTML:{__html:N(r.text)}})]},r.id))})]})]})}function _e({comments:e}){return e.length===0?null:m("div",{class:"orphaned-comments-section",children:[m("h2",{children:"Unmapped Comments"}),m("p",{class:"orphaned-comments-note",children:"The following comments could not be matched to any current paragraph. They may refer to content that has been removed or significantly changed."}),m("div",{class:"orphaned-comments-list",children:e.map(t=>m(nt,{comment:t},t.id))})]})}var qe={similarityThreshold:.85,orphanedLocation:"end-of-chapter",showCommentCount:!0},d={backend:null,allComments:[],currentPageComments:[],orphanedComments:[]};async function ot(e){d.backend=e,console.log("Initializing mdbook-comments base module..."),d.backend.init&&await d.backend.init(),"onAuthChange"in d.backend&&d.backend.onAuthChange&&d.backend.onAuthChange(()=>{ft()}),await I(),ht()}async function I(){if(!d.backend)throw new Error("Backend not initialized");try{d.allComments=await d.backend.loadComments(),console.log(`Loaded ${d.allComments.length} comments`),rt(),at(),ct(),lt()}catch(e){console.error("Error loading comments:",e)}}function rt(){let e={};d.allComments.forEach(t=>{t.replies=[],e[t.id]=t}),d.allComments.forEach(t=>{if(t.parent_id&&e[t.parent_id]){let n=e[t.parent_id];n&&n.replies&&n.replies.push(t)}})}function at(){d.currentPageComments=[],d.orphanedComments=[];let e=document.querySelectorAll(".comment-link-wrapper"),t=new Set;e.forEach(n=>{let o=n.getAttribute("data-comment-id");if(!o)return;d.allComments.filter(r=>r.metadata&&r.metadata.id===o&&!t.has(r.id)&&!r.parent_id).forEach(r=>{d.currentPageComments.push({paragraphId:o,comment:r,confidence:1}),t.add(r.id)})}),e.forEach(n=>{let o=n.getAttribute("data-comment-id"),a=n.getAttribute("data-comment-meta")||"{}",r=JSON.parse(a);o&&d.allComments.forEach(i=>{if(t.has(i.id)||!i.metadata||i.parent_id)return;let l=it(r,i.metadata);l>=qe.similarityThreshold&&(d.currentPageComments.push({paragraphId:o,comment:i,confidence:l}),t.add(i.id))})}),d.allComments.forEach(n=>{!t.has(n.id)&&!n.parent_id&&d.orphanedComments.push(n)}),console.log(`Matched ${d.currentPageComments.length} comments, ${d.orphanedComments.length} orphaned`)}function it(e,t){let n=0,o=0;if(e.content&&t.content){let a=ue(e.content,t.content);n+=a*.5,o+=.5}if(e.context&&t.context){if(e.context.prev&&t.context.prev){let a=ue(e.context.prev,t.context.prev);n+=a*.2,o+=.2}if(e.context.next&&t.context.next){let a=ue(e.context.next,t.context.next);n+=a*.2,o+=.2}if(e.context["heading-path"]&&t.context["heading-path"]){let a=st(e.context["heading-path"],t.context["heading-path"]);n+=a*.1,o+=.1}}return o>0?n/o:0}function ue(e,t){let n=new Set(ze(e)),o=new Set(ze(t)),a=new Set(Array.from(n).filter(i=>o.has(i))),r=new Set([...Array.from(n),...Array.from(o)]);return r.size>0?a.size/r.size:0}function ze(e){return e.toLowerCase().replace(/[^\w\s]/g," ").split(/\s+/).filter(t=>t.length>2)}function st(e,t){let n=new Set(e),o=new Set(t),a=new Set(Array.from(n).filter(i=>o.has(i))),r=new Set([...Array.from(n),...Array.from(o)]);return r.size>0?a.size/r.size:0}function ct(){document.querySelectorAll(".comment-link-wrapper").forEach(e=>{let t=e.getAttribute("data-comment-id");if(!t)return;let n=d.currentPageComments.filter(o=>o.paragraphId===t).length;if(n>0&&qe.showCommentCount){let o=e.querySelector(".comment-link");o&&(o.textContent=`comment (${n})`)}})}function lt(){let e=document.querySelector(".orphaned-comments-section");if(e&&e.remove(),d.orphanedComments.length===0)return;let t=document.querySelector("main")||document.querySelector("#content")||document.body,n=document.createElement("div");t.appendChild(n),$(m(_e,{comments:d.orphanedComments}),n)}function mt(e){let t=document.getElementById(`comments-${e}`);if(t){t.style.display=t.style.display==="none"?"block":"none";return}t=_t(e);let n=document.querySelector(`[data-comment-id="${e}"]`);n&&n.parentNode&&n.parentNode.insertBefore(t,n.nextSibling)}function _t(e){if(!d.backend)throw new Error("Backend not initialized");let n=document.querySelector(`[data-comment-id="${e}"]`)?.getAttribute("data-comment-meta")||"{}",o=JSON.parse(n),a=d.currentPageComments.filter(l=>l.paragraphId===e),r=document.createElement("div");r.style.display="contents";let i=async()=>{await I();let l=d.currentPageComments.filter(_=>_.paragraphId===e);$(m(B,{paragraphId:e,metadata:o,comments:l,backend:d.backend,onUpdate:i}),r)};return $(m(B,{paragraphId:e,metadata:o,comments:a,backend:d.backend,onUpdate:i}),r),r}function ut(e){let t=document.getElementById(`reply-form-${e}`);t&&(t.style.display=t.style.display==="none"?"block":"none")}async function dt(e){if(!d.backend)throw new Error("Backend not initialized");let t=document.querySelector(`[data-comment-id="${e}"]`);if(!t)return;let n=document.getElementById(`comments-${e}`);if(!n)return;let o=n.querySelector(".comment-form");if(!o)return;let a=d.backend.getCurrentAuthor?d.backend.getCurrentAuthor():null;if(d.backend.showAuthorInput&&!a){let c=o.querySelector(".author-input");if(c){if(a=c.value.trim(),!a){alert("Please enter your name");return}d.backend.setCurrentAuthor&&d.backend.setCurrentAuthor(a)}}let r=o.querySelector(".comment-input");if(!r)return;let i=r.value.trim();if(!i){alert("Please enter a comment");return}let l=t.getAttribute("data-comment-meta")||"{}",_=JSON.parse(l);try{let c=await d.backend.saveComment(e,_,i,a||"Anonymous");c.replies=c.replies||[],d.allComments.push(c),d.currentPageComments.push({paragraphId:e,comment:c,confidence:1}),r.value="",await I()}catch(c){console.error("Error posting comment:",c),alert("Failed to post comment. Please try again.")}}async function pt(e){if(!d.backend)throw new Error("Backend not initialized");let t=d.backend.getCurrentAuthor?d.backend.getCurrentAuthor():null,n=document.getElementById(`reply-form-${e}`);if(!n)return;let o=n.querySelector(".reply-input");if(!o)return;let a=o.value.trim();if(!a){alert("Please enter a reply");return}try{let r=await d.backend.saveReply(e,a,t||"Anonymous"),i=d.allComments.find(l=>l.id===e);i&&(i.replies||(i.replies=[]),i.replies.push(r)),d.allComments.push(r),o.value="",n.style.display="none",await I()}catch(r){console.error("Error posting reply:",r),alert("Failed to post reply. Please try again.")}}async function ft(){await I(),document.querySelectorAll(".comment-section").forEach(e=>{let t=e.getAttribute("data-paragraph-id");if(t&&e.style.display!=="none"){let o=e.parentElement;if(o&&d.backend){let r=document.querySelector(`[data-comment-id="${t}"]`)?.getAttribute("data-comment-meta")||"{}",i=JSON.parse(r),l=d.currentPageComments.filter(_=>_.paragraphId===t);$(m(B,{paragraphId:t,metadata:i,comments:l,backend:d.backend,onUpdate:()=>I()}),o,e)}}})}function Gt(e){let t=document.createElement("div");return t.textContent=e,t.innerHTML}function Kt(e){return e?new Date(e).toLocaleString():""}function ht(){if(document.getElementById("mdbook-comments-styles"))return;let e=document.createElement("style");e.id="mdbook-comments-styles",e.textContent=`
            .comment-link-wrapper {
                display: inline;
                margin-left: 0.5em;
            }

            .comment-link {
                font-size: 0.85em;
                color: #0066cc;
                text-decoration: underline;
                cursor: pointer;
            }

            .comment-link:hover {
                color: #0052a3;
            }

            .comment-section {
                margin: 1em 0;
                padding: 1em;
                background: #f5f5f5;
                border-left: 3px solid #0066cc;
                border-radius: 4px;
            }

            .comment-list {
                margin-bottom: 1em;
            }

            .comment-item {
                background: white;
                padding: 0.75em;
                margin-bottom: 0.75em;
                border-radius: 4px;
                box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            }

            .comment-header {
                display: flex;
                justify-content: space-between;
                margin-bottom: 0.5em;
                font-size: 0.9em;
                color: #666;
            }

            .comment-author {
                font-weight: bold;
                color: #333;
            }

            .comment-text {
                line-height: 1.5;
                white-space: pre-wrap;
            }

            .comment-replies {
                margin-top: 0.75em;
                margin-left: 1.5em;
                border-left: 2px solid #ddd;
                padding-left: 0.75em;
            }

            .reply-item {
                background: #fafafa;
                padding: 0.5em;
                margin-bottom: 0.5em;
                border-radius: 3px;
            }

            .reply-header {
                display: flex;
                justify-content: space-between;
                margin-bottom: 0.25em;
                font-size: 0.85em;
                color: #666;
            }

            .reply-author {
                font-weight: bold;
                color: #333;
            }

            .reply-text {
                font-size: 0.95em;
                line-height: 1.4;
                white-space: pre-wrap;
            }

            .comment-reply-btn {
                margin-top: 0.5em;
                padding: 0.25em 0.75em;
                font-size: 0.85em;
                background: #f0f0f0;
                border: 1px solid #ddd;
                border-radius: 3px;
                cursor: pointer;
            }

            .comment-reply-btn:hover {
                background: #e0e0e0;
            }

            .comment-form, .reply-form {
                margin-top: 0.75em;
            }

            .author-input {
                width: 100%;
                padding: 0.5em;
                margin-bottom: 0.5em;
                border: 1px solid #ddd;
                border-radius: 3px;
                font-family: inherit;
                font-size: 0.95em;
            }

            .comment-input, .reply-input {
                width: 100%;
                padding: 0.5em;
                border: 1px solid #ddd;
                border-radius: 3px;
                font-family: inherit;
                font-size: 0.95em;
                resize: vertical;
            }

            .comment-submit, .reply-submit {
                margin-top: 0.5em;
                padding: 0.5em 1em;
                background: #0066cc;
                color: white;
                border: none;
                border-radius: 3px;
                cursor: pointer;
                font-size: 0.95em;
            }

            .comment-submit:hover, .reply-submit:hover {
                background: #0052a3;
            }

            .no-comments {
                color: #999;
                font-style: italic;
            }

            .orphaned-comments-section {
                margin-top: 3em;
                padding-top: 2em;
                border-top: 2px solid #ddd;
            }

            .orphaned-comments-note {
                color: #666;
                font-style: italic;
                margin-bottom: 1.5em;
            }

            .orphaned-comment {
                margin-bottom: 2em;
                padding: 1em;
                background: #fff9e6;
                border-left: 3px solid #ffcc00;
                border-radius: 4px;
            }

            .orphaned-comment-context {
                margin-bottom: 1em;
                padding-bottom: 1em;
                border-bottom: 1px solid #ddd;
            }

            .orphaned-comment-context blockquote {
                margin: 0.5em 0;
                padding: 0.5em;
                background: white;
                border-left: 3px solid #ddd;
                font-style: italic;
            }

            .orphaned-comment-location {
                font-size: 0.9em;
                color: #666;
                margin-top: 0.5em;
            }
        `,document.head.appendChild(e)}window.MdbookComments={init:ot};window.toggleComments=mt;window.submitComment=dt;window.submitReply=pt;window.showReplyForm=ut;})();
//# sourceMappingURL=comments-base.js.map
