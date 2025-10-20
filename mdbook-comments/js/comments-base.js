"use strict";(()=>{var y={similarityThreshold:.85,orphanedLocation:"end-of-chapter",showCommentCount:!0},r={backend:null,allComments:[],currentPageComments:[],orphanedComments:[]};async function w(e){r.backend=e,console.log("Initializing mdbook-comments base module..."),r.backend.init&&await r.backend.init(),"onAuthChange"in r.backend&&r.backend.onAuthChange&&r.backend.onAuthChange(()=>{q()}),await E(),B()}async function E(){if(!r.backend)throw new Error("Backend not initialized");try{r.allComments=await r.backend.loadComments(),console.log(`Loaded ${r.allComments.length} comments`),A(),S(),$(),L()}catch(e){console.error("Error loading comments:",e)}}function A(){let e={};r.allComments.forEach(t=>{t.replies||(t.replies=[]),e[t.id]=t}),r.allComments.forEach(t=>{if(t.parent_id&&e[t.parent_id]){let n=e[t.parent_id];n&&n.replies&&n.replies.push(t)}})}function S(){r.currentPageComments=[],r.orphanedComments=[];let e=document.querySelectorAll(".comment-link-wrapper"),t=new Set;e.forEach(n=>{let o=n.getAttribute("data-comment-id");if(!o)return;r.allComments.filter(a=>a.metadata&&a.metadata.id===o&&!t.has(a.id)&&!a.parent_id).forEach(a=>{r.currentPageComments.push({paragraphId:o,comment:a,confidence:1}),t.add(a.id)})}),e.forEach(n=>{let o=n.getAttribute("data-comment-id"),i=n.getAttribute("data-comment-meta")||"{}",a=JSON.parse(i);o&&r.allComments.forEach(m=>{if(t.has(m.id)||!m.metadata||m.parent_id)return;let s=M(a,m.metadata);s>=y.similarityThreshold&&(r.currentPageComments.push({paragraphId:o,comment:m,confidence:s}),t.add(m.id))})}),r.allComments.forEach(n=>{!t.has(n.id)&&!n.parent_id&&r.orphanedComments.push(n)}),console.log(`Matched ${r.currentPageComments.length} comments, ${r.orphanedComments.length} orphaned`)}function M(e,t){let n=0,o=0;if(e.content&&t.content){let i=u(e.content,t.content);n+=i*.5,o+=.5}if(e.context&&t.context){if(e.context.prev&&t.context.prev){let i=u(e.context.prev,t.context.prev);n+=i*.2,o+=.2}if(e.context.next&&t.context.next){let i=u(e.context.next,t.context.next);n+=i*.2,o+=.2}if(e.context["heading-path"]&&t.context["heading-path"]){let i=P(e.context["heading-path"],t.context["heading-path"]);n+=i*.1,o+=.1}}return o>0?n/o:0}function u(e,t){let n=new Set(g(e)),o=new Set(g(t)),i=new Set(Array.from(n).filter(m=>o.has(m))),a=new Set([...Array.from(n),...Array.from(o)]);return a.size>0?i.size/a.size:0}function g(e){return e.toLowerCase().replace(/[^\w\s]/g," ").split(/\s+/).filter(t=>t.length>2)}function P(e,t){let n=new Set(e),o=new Set(t),i=new Set(Array.from(n).filter(m=>o.has(m))),a=new Set([...Array.from(n),...Array.from(o)]);return a.size>0?i.size/a.size:0}function $(){document.querySelectorAll(".comment-link-wrapper").forEach(e=>{let t=e.getAttribute("data-comment-id");if(!t)return;let n=r.currentPageComments.filter(o=>o.paragraphId===t).length;if(n>0&&y.showCommentCount){let o=e.querySelector(".comment-link");o&&(o.textContent=`comment (${n})`)}})}function L(){let e=document.querySelector(".orphaned-comments-section");if(e&&e.remove(),r.orphanedComments.length===0)return;let t=document.querySelector("main")||document.querySelector("#content")||document.body,n=document.createElement("div");n.className="orphaned-comments-section",n.innerHTML=`
            <h2>Unmapped Comments</h2>
            <p class="orphaned-comments-note">
                The following comments could not be matched to any current paragraph.
                They may refer to content that has been removed or significantly changed.
            </p>
            <div class="orphaned-comments-list"></div>
        `;let o=n.querySelector(".orphaned-comments-list");o&&(r.orphanedComments.forEach(i=>{let a=T(i);o.appendChild(a)}),t.appendChild(n))}function T(e){let t=document.createElement("div");t.className="orphaned-comment";let n=e.metadata||{},o=n.context||{"heading-path":[]},i=n.content||"[Content not available]";return t.innerHTML=`
            <div class="orphaned-comment-context">
                <strong>Original paragraph:</strong>
                <blockquote>${l(i)}</blockquote>
                ${o["heading-path"]&&o["heading-path"].length>0?`
                    <div class="orphaned-comment-location">
                        Section: ${o["heading-path"].join(" > ")}
                    </div>
                `:""}
            </div>
            <div class="comment-item">
                <div class="comment-header">
                    <span class="comment-author">${l(e.author||"Anonymous")}</span>
                    <span class="comment-date">${h(e.created)}</span>
                </div>
                <div class="comment-text">${l(e.text)}</div>
                ${e.replies&&e.replies.length>0?`
                    <div class="comment-replies">
                        ${e.replies.map(a=>C(a)).join("")}
                    </div>
                `:""}
            </div>
        `,t}function p(e){let t=document.getElementById(`comments-${e}`);if(t){t.style.display=t.style.display==="none"?"block":"none";return}t=z(e);let n=document.querySelector(`[data-comment-id="${e}"]`);n&&n.parentNode&&n.parentNode.insertBefore(t,n.nextSibling)}function z(e){let t=document.createElement("div");t.id=`comments-${e}`,t.className="comment-section",t.setAttribute("data-paragraph-id",e);let n=r.currentPageComments.filter(a=>a.paragraphId===e).map(a=>a.comment),o=document.createElement("div");o.className="comment-list",n.length>0?n.forEach(a=>{let m=H(a);o.appendChild(m)}):o.innerHTML='<p class="no-comments">No comments yet. Be the first to comment!</p>',t.appendChild(o);let i=b(e);return t.appendChild(i),t}function b(e,t=null){if(!r.backend)throw new Error("Backend not initialized");let n=document.createElement("div");n.className=t?"reply-form":"comment-form";let o=!!t,i=r.backend.getCurrentAuthor?r.backend.getCurrentAuthor():"";if(r.backend.showAuthorInput&&!i){let s=document.createElement("input");s.type="text",s.className="author-input",s.name="author",s.placeholder="Your name",s.value=i||"",s.addEventListener("input",d=>{let f=d.target.value.trim();f&&r.backend&&r.backend.setCurrentAuthor&&r.backend.setCurrentAuthor(f)}),n.appendChild(s)}let a=document.createElement("textarea");a.className=o?"reply-input":"comment-input",a.name=o?"reply-text":"comment-text",a.placeholder=o?"Write a reply...":"Add a comment...",a.rows=o?2:3,n.appendChild(a);let m=document.createElement("button");return m.type="submit",m.className=o?"reply-submit":"comment-submit",m.textContent=o?"Post Reply":"Submit",m.onclick=s=>{s.preventDefault(),o&&t?k(t):e&&v(e)},n.appendChild(m),n}function H(e){let t=document.createElement("div");if(t.className="comment-item",t.setAttribute("data-comment-id",e.id),t.innerHTML=`
            <div class="comment-header">
                <span class="comment-author">${l(e.author||"Anonymous")}</span>
                <span class="comment-date">${h(e.created)}</span>
            </div>
            <div class="comment-text">${l(e.text)}</div>
        `,e.replies&&e.replies.length>0){let i=document.createElement("div");i.className="comment-replies",e.replies.forEach(a=>{i.innerHTML+=C(a)}),t.appendChild(i)}let n=document.createElement("button");n.className="comment-reply-btn",n.textContent="Reply",n.onclick=()=>x(e.id),t.appendChild(n);let o=b(null,e.id);return o.id=`reply-form-${e.id}`,o.style.display="none",t.appendChild(o),t}function C(e){return`
            <div class="reply-item">
                <div class="reply-header">
                    <span class="reply-author">${l(e.author||"Anonymous")}</span>
                    <span class="reply-date">${h(e.created)}</span>
                </div>
                <div class="reply-text">${l(e.text)}</div>
            </div>
        `}function x(e){let t=document.getElementById(`reply-form-${e}`);t&&(t.style.display=t.style.display==="none"?"block":"none")}async function v(e){if(!r.backend)throw new Error("Backend not initialized");let t=document.querySelector(`[data-comment-id="${e}"]`);if(!t)return;let n=document.getElementById(`comments-${e}`);if(!n)return;let o=n.querySelector(".comment-form");if(!o)return;let i=r.backend.getCurrentAuthor?r.backend.getCurrentAuthor():null;if(r.backend.showAuthorInput&&!i){let c=o.querySelector(".author-input");if(c){if(i=c.value.trim(),!i){alert("Please enter your name");return}r.backend.setCurrentAuthor&&r.backend.setCurrentAuthor(i)}}let a=o.querySelector(".comment-input");if(!a)return;let m=a.value.trim();if(!m){alert("Please enter a comment");return}let s=t.getAttribute("data-comment-meta")||"{}",d=JSON.parse(s);try{let c=await r.backend.saveComment(e,d,m,i||"Anonymous");c.replies=c.replies||[],r.allComments.push(c),r.currentPageComments.push({paragraphId:e,comment:c,confidence:1}),a.value="",n.remove(),p(e)}catch(c){console.error("Error posting comment:",c),alert("Failed to post comment. Please try again.")}}async function k(e){if(!r.backend)throw new Error("Backend not initialized");let t=r.backend.getCurrentAuthor?r.backend.getCurrentAuthor():null,n=document.getElementById(`reply-form-${e}`);if(!n)return;let o=n.querySelector(".reply-input");if(!o)return;let i=o.value.trim();if(!i){alert("Please enter a reply");return}let a=n.closest(".comment-section"),m=a?a.getAttribute("data-paragraph-id"):null;try{let s=await r.backend.saveReply(e,i,t||"Anonymous"),d=r.allComments.find(c=>c.id===e);d&&(d.replies||(d.replies=[]),d.replies.push(s)),r.allComments.push(s),o.value="",n.style.display="none",a&&m&&(a.remove(),p(m))}catch(s){console.error("Error posting reply:",s),alert("Failed to post reply. Please try again.")}}function q(){document.querySelectorAll(".comment-section").forEach(e=>{let t=e.getAttribute("data-paragraph-id");t&&e.style.display!=="none"&&(e.remove(),p(t))})}function l(e){let t=document.createElement("div");return t.textContent=e,t.innerHTML}function h(e){return e?new Date(e).toLocaleString():""}function B(){if(document.getElementById("mdbook-comments-styles"))return;let e=document.createElement("style");e.id="mdbook-comments-styles",e.textContent=`
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
        `,document.head.appendChild(e)}window.MdbookComments={init:w};window.toggleComments=p;window.submitComment=v;window.submitReply=k;window.showReplyForm=x;})();
//# sourceMappingURL=comments-base.js.map
